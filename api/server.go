package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/maybemaby/oapibase/api/auth"
)

type Server struct {
	logger *slog.Logger
	port   string
	srv    *http.Server
	db     *sql.DB
	// dbx *sqlx.DB
	pool       *pgxpool.Pool
	services   *services
	jwtManager *auth.JwtManager
	prod       bool
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
