package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

const OAUTH_STATE_SESSION_KEY = "oauth_state"
const OAUTH_VERIFIER_SESSION_KEY = "oauth_verifier"

var ErrStateMismatch = errors.New("state mismatch")

// GenerateState generates a random state string, base64 urlencoded with a length of 64 bytes
func GenerateState() (string, error) {
	nonceBytes := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, nonceBytes)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

// ValidateState checks if the state in the query matches the state in the cookie,
// returns ErrStateMismatch if the states do not match
// Assumes the state cookie name is OAUTH_STATE_SESSION_KEY
func ValidateState(r *http.Request) error {
	state := r.URL.Query().Get("state")

	if state == "" {
		return errors.New("missing state")
	}

	cookie, err := r.Cookie(OAUTH_STATE_SESSION_KEY)

	if err != nil {
		return err
	}

	if cookie.Value != state {
		return ErrStateMismatch
	}

	return nil
}

func UpdateAccountTokens(ctx context.Context,
	userId int, provider, accessToken, refreshToken string, accessTokenExpiration, refreshTokenExpiration *time.Time,
	db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `UPDATE accounts SET access_token = $1, refresh_token = $2, access_token_expires_at = $3, refresh_token_expires_at = $4
	WHERE user_id = $5 AND provider = $6
	`, accessToken, refreshToken, accessTokenExpiration, refreshTokenExpiration, userId, provider)

	return err
}

type OAuthProvider struct {
	Config *oauth2.Config
}

func NewOAuthProvider(config *oauth2.Config) *OAuthProvider {
	return &OAuthProvider{
		Config: config,
	}
}

func (p *OAuthProvider) InitStateAndVerifier(w http.ResponseWriter) (string, string, error) {
	state, err := GenerateState()
	if err != nil {
		return "", "", err
	}

	verifier := oauth2.GenerateVerifier()

	http.SetCookie(w, &http.Cookie{
		Name:     OAUTH_STATE_SESSION_KEY,
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     OAUTH_VERIFIER_SESSION_KEY,
		Value:    verifier,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		SameSite: http.SameSiteLaxMode,
	})

	return state, verifier, nil
}
