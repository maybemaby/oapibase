package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maybemaby/oapibase/api/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const userInfoEndpoint = "https://openidconnect.googleapis.com/v1/userinfo"

type GoogleHandler struct {
	Provider   *auth.OAuthProvider
	DB         *pgxpool.Pool
	jwtManager *auth.JwtManager
}

func NewGoogleHandler(db *pgxpool.Pool, jwtManager *auth.JwtManager) *GoogleHandler {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}

	provider := &auth.OAuthProvider{
		Config: config,
	}

	return &GoogleHandler{
		Provider:   provider,
		DB:         db,
		jwtManager: jwtManager,
	}
}

// GoogleToken is a custom struct to hold the oidc token response
// ExpiresIn remaining lifetime of the token in seconds
type GoogleToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
	IDToken      string    `json:"id_token"`
	ExpiresIn    *int      `json:"expires_in,omitempty"`
	Scope        string    `json:"scope"`
}

// For OIDC
type googleUserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Locale        string `json:"locale"`
}

func parseGoogleToken(tok *oauth2.Token) *GoogleToken {

	tokExpiresIn := tok.Extra("expires_in")

	if tokExpiresIn != nil {
		expiresIn := int(math.Round(tokExpiresIn.(float64)))

		return &GoogleToken{
			AccessToken:  tok.AccessToken,
			RefreshToken: tok.RefreshToken,
			Expiry:       tok.Expiry,
			IDToken:      tok.Extra("id_token").(string),
			ExpiresIn:    &expiresIn,
			Scope:        tok.Extra("scope").(string),
		}
	}

	return &GoogleToken{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
		IDToken:      tok.Extra("id_token").(string),
		ExpiresIn:    nil,
		Scope:        tok.Extra("scope").(string),
	}
}

func (h *GoogleHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	// Github does not require a verifier, so we can skip that step
	state, verifier, err := h.Provider.InitStateAndVerifier(w)

	if err != nil {
		http.Error(w, "Failed to initialize state and verifier", http.StatusInternalServerError)
		return
	}

	url := h.Provider.Config.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(verifier))

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *GoogleHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	stateErr := auth.ValidateState(r)

	if stateErr != nil {
		http.Error(w, "State validation failed", http.StatusBadRequest)
		return
	}

	verifierCookie, err := r.Cookie(auth.OAUTH_VERIFIER_SESSION_KEY)

	if err != nil {
		http.Error(w, "Missing verifier cookie", http.StatusBadRequest)
		return
	}

	tok, err := h.Provider.Config.Exchange(r.Context(), code, oauth2.VerifierOption(verifierCookie.Value))

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	googleToken := parseGoogleToken(tok)

	client := h.Provider.Config.Client(r.Context(), tok)

	userInfo, err := client.Get(userInfoEndpoint)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	defer userInfo.Body.Close()

	var user googleUserInfo

	if err := json.NewDecoder(userInfo.Body).Decode(&user); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	existingUser, existingAccount, err := auth.GetUserAccountByEmail(r.Context(), user.Email, "google", h.DB)

	// No error, so we have an existing user or account
	if err == nil {

		status := auth.UserAccountStatus(existingUser, existingAccount)

		switch status {
		case auth.AccountStatusNoAccount:
			// Probably don't want to link accounts by default
			w.Write([]byte("Could not login"))

			return
		case auth.AccountStatusLinked:
			// User already exists and account is linked, refresh tokens and login
			err = auth.UpdateAccountTokens(r.Context(), existingUser.ID, "google",
				googleToken.AccessToken, googleToken.RefreshToken, &googleToken.Expiry, nil, h.DB)

			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			accessToken, err := h.jwtManager.EncodeAccessToken(auth.SessionData{
				UserId: existingUser.ID,
				Role:   existingUser.Role,
			})

			refreshToken, refreshErr := h.jwtManager.EncodeRefreshToken(auth.SessionData{
				UserId: existingUser.ID,
				Role:   existingUser.Role,
			})

			jwtErr := errors.Join(err, refreshErr)

			if jwtErr != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Write([]byte(accessToken + "\n" + refreshToken))
			return
		}
	} else if err != sql.ErrNoRows {
		// Unexpected error, log it and return
		slog.Error("Error during Google login", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	newUser, err := auth.CreateUserAccount(r.Context(), auth.User{
		Email:        &user.Email,
		PasswordHash: nil,
		Role:         "user",
	}, auth.AccountInsert{
		Provider:              "google",
		ProviderId:            user.Sub,
		AccessToken:           googleToken.AccessToken,
		AccessTokenExpiresAt:  googleToken.Expiry,
		RefreshToken:          googleToken.RefreshToken,
		RefreshTokenExpiresAt: nil, // Google does not return this
	}, h.DB)

	if err != nil {
		logger := RequestLogger(r)
		logger.Error("Error creating user account", slog.Any("err", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accessToken, err := h.jwtManager.EncodeAccessToken(auth.SessionData{
		UserId: newUser.ID,
		Role:   newUser.Role,
	})

	refreshToken, refreshErr := h.jwtManager.EncodeRefreshToken(auth.SessionData{
		UserId: newUser.ID,
		Role:   newUser.Role,
	})

	jwtErr := errors.Join(err, refreshErr)

	if jwtErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":    newUser.ID,
			"email": newUser.Email,
			"role":  newUser.Role,
		},
		"token": map[string]interface{}{
			"accessToken":  googleToken.AccessToken,
			"refreshToken": googleToken.RefreshToken,
			"expiresIn":    googleToken.ExpiresIn,
			"expiry":       googleToken.Expiry,
			"scope":        googleToken.Scope,
		},
		"jwt": map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
