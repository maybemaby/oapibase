package api

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maybemaby/smolauth"
)

type features struct {
	twitchEnabled  bool
	youtubeEnabled bool
}

type services struct {

}

func newServices(pool *pgxpool.Pool, logger *slog.Logger, authManager *smolauth.AuthManager) (*services) {


	return &services{

		}
}
