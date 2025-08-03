package api

import (
	"log/slog"
	"net/http"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/option"
)

func SpecRouter() spec.Generator {
	r := spec.NewRouter(
		option.WithTitle("oapibase"),
		option.WithVersion("0.1.0"),
	)

	r.Post("/auth/login",
		option.Request(new(PassLoginBody)),
		option.Response(200, "Success"),
		option.Tags("auth"),
	)

	r.Post("/auth/signup", option.Request(new(PassSignupBody)),
		option.Response(200, "Success"),
		option.Tags("auth"),
	)

	r.Get("/auth/me",
		option.Response(200, new(MeResponse)),
		option.Response(401, "Unauthorized"),
		option.Tags("auth"),
	)

	return r
}

func GenSpec() ([]byte, error) {
	r := SpecRouter()

	return r.GenerateSchema("json")
}

func HandleSpec(w http.ResponseWriter, r *http.Request) {
	logger := RequestLogger(r)
	swagger, err := GenSpec()

	if err != nil {
		logger.Error("Failed to generate Swagger spec", slog.Any("err", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(swagger)
}
