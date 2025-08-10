package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int       `json:"id"`
	Email        *string   `json:"email"`
	Role         string    `json:"role"`
	PasswordHash *string   `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func GetUserByEmail(ctx context.Context, email string, db *pgxpool.Pool) (User, error) {
	var user User

	err := db.QueryRow(ctx, "SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func CreateUser(ctx context.Context, email, password string, db *pgxpool.Pool) (User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	row := db.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at", email, hashedPassword)

	var id int
	var createdAt time.Time
	err = row.Scan(&id, &createdAt)

	if err != nil {
		return User{}, err
	}

	return User{
		ID:           int(id),
		Email:        &email,
		PasswordHash: &hashedPassword,
		CreatedAt:    createdAt,
		Role:         "user",
	}, nil
}

