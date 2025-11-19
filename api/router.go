package api

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/maybemaby/oapibase/api/auth"
	"github.com/oaswrap/spec-ui/config"
	"github.com/oaswrap/spec/adapter/httpopenapi"
	"github.com/oaswrap/spec/option"
)

func (s *Server) MountRoutesOapi() {

	mux := http.NewServeMux()

	authHandler := &AuthHandler{
		authManager: s.authManager,
		jwtManager:  s.jwtManager,
		pool:        s.pool,
	}

	googleHandler := NewGoogleHandler(s.pool, s.jwtManager)

	rootMw := RootMiddleware(s.logger, MiddlewareConfig{
		CorsOrigin: "http://localhost:3001",
	})

	authMw := rootMw.Append(auth.RequireAccessToken(s.jwtManager))

	r := httpopenapi.NewGenerator(mux,
		option.WithTitle("oapibase"),
		option.WithVersion("0.1.0"),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("Bearer")),
		option.WithSwaggerUI(config.SwaggerUI{
			UIConfig: map[string]string{
				"persistAuthorization": "true",
			},
		}),
		option.WithExternalDocs("/docs/openapi.json"),
		option.WithDisableDocs(s.prod),
	)

	if !s.prod {

		s.logger.Info("Swagger docs enabled at /docs")

		r.HandleFunc("/docs/openapi.json", func(w http.ResponseWriter, req *http.Request) {
			spec, err := r.MarshalJSON()

			if err != nil {
				http.Error(w, "Error generating OpenAPI spec", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(spec)
		})
	}

	authRoute := r.Group("/auth").With(option.GroupTags("auth"))

	authRoute.Handle("GET /me", authMw.ThenFunc(authHandler.GetAuthMe)).With(
		option.Response(200, new(MeResponse)),
		option.Response(401, "Unauthorized"),
	)

	authRoute.Handle("POST /signup", rootMw.ThenFunc(authHandler.SignupJWT)).With(
		option.Request(new(PassSignupBody)),
		ResponsesWithDefault(map[int]any{
			200: new(LoginJwtResponse),
		}),
	)

	authRoute.Handle("POST /login", rootMw.ThenFunc(authHandler.LoginJWT)).With(
		option.Request(new(PassLoginBody)),
		Responses(map[int]any{
			401: new(AuthErrorResponse),
			200: new(LoginJwtResponse),
		}),
	)

	authRoute.Handle("GET /google", rootMw.ThenFunc(googleHandler.HandleAuth))
	authRoute.Handle("GET /google/callback", rootMw.ThenFunc(googleHandler.HandleCallback))

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	s.srv = srv
}

func MountSpa(mux *http.ServeMux, pattern string, filesys fs.FS) {
	fileServer := http.FileServer(http.FS(filesys))

	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to open the requested file
		file, err := filesys.Open(path)
		if err == nil {
			// File exists, close it and let the file server handle it
			file.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist, check if it's a not found error
		if os.IsNotExist(err) {
			// Serve index.html as fallback for SPA routing
			indexData, err := fs.ReadFile(filesys, "index.html")
			if err != nil {
				http.Error(w, "Index file not found", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(indexData)
			return
		}

		// Other error occurred
		http.Error(w, "Error accessing file", http.StatusInternalServerError)
	})
}
