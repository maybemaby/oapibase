package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JwtManager struct {
	AccessTokenSecret    []byte
	RefreshTokenSecret   []byte
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
}

func (m *JwtManager) EncodeAccessToken(data SessionData) (string, error) {
	claims := AccessTokenClaims{
		UserId: data.UserId,
		Role:   data.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.AccessTokenLifetime)),
			Subject:   strconv.Itoa(data.UserId),
			Issuer:    "go-auth-snippets",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.AccessTokenSecret)
}

func (m *JwtManager) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return m.AccessTokenSecret, nil
	}, jwt.WithLeeway(1*time.Minute))

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)

	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

func RequireAccessToken(manager *JwtManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			parts := strings.Split(token, " ")

			if len(parts) < 2 || parts[1] == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := manager.ValidateAccessToken(parts[1])

			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store the user ID and role in the context
			ctx := context.WithValue(r.Context(), SessionUserIdKey, claims.UserId)
			ctx = context.WithValue(ctx, SessionRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestUser(r *http.Request) (SessionData, error) {
	userId := r.Context().Value(SessionUserIdKey)
	role := r.Context().Value(SessionRoleKey)

	if userId == nil || role == nil {
		return SessionData{}, errors.New("unauthorized")
	}

	return SessionData{
		UserId: userId.(int),
		Role:   role.(string),
	}, nil
}

func (m *JwtManager) EncodeRefreshToken(data SessionData) (string, error) {
	claims := RefreshTokenClaims{
		UserId: data.UserId,
		Role:   data.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.RefreshTokenLifetime)),
			Subject:   strconv.Itoa(data.UserId),
			Issuer:    "go-auth-snippets",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.RefreshTokenSecret)
}

func (m *JwtManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return m.RefreshTokenSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)

	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func RefreshTokenHandler(manager *JwtManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		parts := strings.Split(token, " ")

		if len(parts) < 2 || parts[1] == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := manager.ValidateRefreshToken(parts[1])

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		newAccessToken, err := manager.EncodeAccessToken(SessionData{
			UserId: claims.UserId,
			Role:   claims.Role,
		})

		newRefreshToken, refreshErr := manager.EncodeRefreshToken(SessionData{
			UserId: claims.UserId,
			Role:   claims.Role,
		})

		if errors.Join(err, refreshErr) != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := RefreshTokenResponse{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}
