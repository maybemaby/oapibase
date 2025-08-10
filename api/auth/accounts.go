package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountStatus string

const (
	AccountStatusNoUser    AccountStatus = "no-user"
	AccountStatusNoAccount AccountStatus = "no-account"
	AccountStatusLinked    AccountStatus = "linked"
)

type AccountInsert struct {
	UserId                int        `json:"user_id"`
	Provider              string     `json:"provider"`
	ProviderId            string     `json:"provider_id"`
	AccessToken           string     `json:"access_token"`
	RefreshToken          string     `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time  `json:"access_token_expires_at"`
	RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires_at"`
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

func GetUserAccountByEmail(ctx context.Context, email string, provider string, pool *pgxpool.Pool) (*User, *AccountSelect, error) {
	var user User
	var account AccountSelect

	row := pool.QueryRow(ctx, `SELECT u.id, u.email, a.provider, a.provider_id
	FROM users u
	JOIN accounts a ON u.id = a.user_id
	WHERE u.email = $1 AND a.provider = $2`, email, provider)

	err := row.Scan(&user.ID, &user.Email, &account.Provider, &account.ProviderId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &User{}, &AccountSelect{}, nil // No user found
		}
		return &User{}, &AccountSelect{}, err // Other error
	}

	return &user, &account, nil
}

func UserAccountStatus(user *User, account *AccountSelect) AccountStatus {
	if user.ID == 0 {
		return AccountStatusNoUser
	}
	if account.ProviderId == "" {
		return AccountStatusNoAccount
	}
	return AccountStatusLinked
}

func CreateUserAccount(ctx context.Context, user User, account AccountInsert, db *pgxpool.Pool) (User, error) {

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return User{}, err
	}

	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at", user.Email, user.PasswordHash)

	var id int
	var createdAt time.Time
	err = row.Scan(&id, &createdAt)

	if err != nil {
		return User{}, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO accounts
	(user_id, provider, provider_id, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at)
	 VALUES ($1, $2, $3, $4, $5, $6, $7)`, id, account.Provider, account.ProviderId,
		account.AccessToken, account.AccessTokenExpiresAt,
		account.RefreshToken, account.RefreshTokenExpiresAt)

	if err != nil {
		return User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}

	return User{
		ID:           int(id),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    createdAt,
		Role:         "user",
	}, nil
}
