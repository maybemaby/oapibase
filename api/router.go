package api

import (
	"net/http"

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
		option.WithDisableDocs(s.prod),
	)

	authRoute := r.Group("/auth")

	authRoute.Handle("GET /me", authMw.ThenFunc(authHandler.GetAuthMe)).With(
		option.Response(200, new(MeResponse)),
		option.Response(401, "Unauthorized"),
		option.Tags("auth"),
	)

	authRoute.Handle("POST /signup", rootMw.ThenFunc(authHandler.SignupJWT)).With(
		option.Request(new(PassSignupBody)),
		option.Response(200, new(LoginJwtResponse)),
		option.Tags("auth"),
	)

	authRoute.Handle("POST /login", rootMw.ThenFunc(authHandler.LoginJWT)).With(
		option.Request(new(PassLoginBody)),
		option.Response(200, new(LoginJwtResponse)),
		option.Tags("auth"),
	)

	authRoute.Handle("GET /google", rootMw.ThenFunc(googleHandler.HandleAuth)).With(
		option.Tags("auth"),
	)
	authRoute.Handle("GET /google/callback", rootMw.ThenFunc(googleHandler.HandleCallback)).With(
		option.Tags("auth"),
	)

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	s.srv = srv
}
