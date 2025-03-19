package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountInsert struct {
	UserId               int       `json:"user_id"`
	Provider             string    `json:"provider"`
	ProviderId           string    `json:"provider_id"`
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

type AccountSelect struct {
	Id                   int       `json:"id"`
	UserId               int       `json:"user_id"`
	Provider             string    `json:"provider"`
	ProviderId           string    `json:"provider_id"`
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	CreatedAt            time.Time `json:"created_at"`
}

const insertAccoutSql = `
INSERT into accounts (user_id, provider, provider_id, access_token, refresh_token, access_token_expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (provider, provider_id) DO UPDATE SET access_token = $4, refresh_token = $5, access_token_expires_at = $6
RETURNING id
`

func UpsertAccount(ctx context.Context, pool *pgxpool.Pool, account AccountInsert) (int, error) {

	var id int

	row := pool.QueryRow(ctx, insertAccoutSql,
		account.UserId,
		account.Provider,
		account.ProviderId,
		account.AccessToken,
		account.RefreshToken,
		account.AccessTokenExpiresAt,
	)

	err := row.Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetAccount(ctx context.Context, pool *pgxpool.Pool, provider string, userId int) (AccountSelect, error) {

	rows, err := pool.Query(ctx, "SELECT * FROM accounts WHERE provider = $1 AND user_id = $2 LIMIT 1", provider, userId)

	if err != nil {
		return AccountSelect{}, err
	}

	defer rows.Close()

	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccountSelect])
}

func GetAccountById(ctx context.Context, pool *pgxpool.Pool, id int) (AccountSelect, error) {

	rows, err := pool.Query(ctx, "SELECT * FROM accounts WHERE id = $1 LIMIT 1", id)

	if err != nil {
		return AccountSelect{}, err
	}

	defer rows.Close()

	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccountSelect])
}
