package api

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/maybemaby/oapibase/api/auth"
	"github.com/maybemaby/oapibase/api/utils"
	"github.com/maybemaby/oapibase/frontend"
	"github.com/maybemaby/smolauth"
)

type Server struct {
	logger *slog.Logger
	port   string
	srv    *http.Server
	db     *sql.DB
	// dbx *sqlx.DB
	pool        *pgxpool.Pool
	services    *services
	authManager *smolauth.AuthManager
	jwtManager  *auth.JwtManager
	prod        bool
}

func NewServer(isProd bool) (*Server, error) {

	server := &Server{
		port: "8000",
		prod: isProd,
	}

	server.WithLogger(isProd)

	pool, err := NewPool(context.Background(), !isProd)

	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)

	server.db = db
	server.pool = pool

	jwtManager := &auth.JwtManager{
		AccessTokenSecret:    []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		RefreshTokenSecret:   []byte(os.Getenv("REFRESH_TOKEN_SECRET")),
		AccessTokenLifetime:  time.Minute * 15,
		RefreshTokenLifetime: time.Hour * 24 * 30,
	}

	server.jwtManager = jwtManager

	services := newServices(pool, server.logger, jwtManager)
	server.services = services

	return server, nil
}

func (s *Server) MountRoutes() {

	authHandler := &AuthHandler{
		authManager: s.authManager,
		jwtManager:  s.jwtManager,
		pool:        s.pool,
	}

	googleHandler := NewGoogleHandler(s.pool, s.jwtManager)

	mux := http.NewServeMux()

	rootMw := RootMiddleware(s.logger, MiddlewareConfig{
		CorsOrigin: "http://localhost:3001",
	})

	authMw := rootMw.Append(auth.RequireAccessToken(s.jwtManager))

	if !s.prod {
		mux.HandleFunc("GET /docs/swagger", func(w http.ResponseWriter, r *http.Request) {
			utils.RenderSwaggerUI(w, "/docs/swagger.json")
		})

		s.logger.Info("Swagger UI available at /docs/swagger")
	}

	mux.Handle("GET /auth/me", authMw.ThenFunc(authHandler.GetAuthMe))
	mux.Handle("POST /auth/signup", rootMw.ThenFunc(authHandler.SignupJWT))
	mux.Handle("POST /auth/login", rootMw.ThenFunc(authHandler.LoginJWT))
	mux.Handle("GET /auth/google", rootMw.ThenFunc(googleHandler.HandleAuth))
	mux.Handle("GET /auth/google/callback", rootMw.ThenFunc(googleHandler.HandleCallback))

	// Due to the way the generated code is structured, we need to handle OPTIONS requests explicitly
	mux.Handle("OPTIONS /", rootMw.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	mux.Handle("GET /", rootMw.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		fs, err := fs.Sub(frontend.Assets, "dist")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// file, err := fs.Open(path)
		file, err := frontend.Assets.Open(filepath.Join("dist", path))

		if os.IsNotExist(err) {
			index, err := frontend.Assets.ReadFile("dist/index.html")

			if err != nil {
				s.logger.Error("Failed to get index.html", slog.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(index)
			return
		} else if err != nil {
			s.logger.Error("Failed to get sub directory", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer file.Close()

		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
	}))

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	s.srv = srv
}

func (s *Server) Start(ctx context.Context) error {

	s.MountRoutesOapi()

	s.logger.Info("Server started at http://localhost:" + s.port)
	s.logger.Info(fmt.Sprintf("Server is running in production mode: %t", s.prod))
	s.logger.Debug("Server is running in debug mode")

	return s.srv.ListenAndServe()
}

func (s *Server) WithLogger(isProd bool) {
	format := JSONFormat
	level := slog.LevelInfo

	if !isProd {
		level = slog.LevelDebug
		format = TEXTFormat
	}

	s.logger = BootstrapLogger(level, format, !isProd)
}

func (s *Server) WithPort(port string) {
	s.port = port
}
