package api

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maybemaby/oapibase/api/auth"
)

type services struct {
}

func newServices(pool *pgxpool.Pool, logger *slog.Logger, authManager *auth.JwtManager) *services {

	return &services{}
}
