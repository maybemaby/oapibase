package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maybemaby/oapibase/api/auth"
	"github.com/maybemaby/oapibase/api/utils"
)

type AuthHandler struct {
	jwtManager *auth.JwtManager
	pool       *pgxpool.Pool
}

type PassLoginBody struct {
	Email    string `json:"email" example:"email@site.com"`
	Password string `json:"password"`
}

type PassSignupBody struct {
	Email     string `json:"email" example:"email@site.com"`
	Password  string `json:"password" minLength:"8"`
	Password2 string `json:"password2"`
}

type LoginJwtResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) SignupJWT(w http.ResponseWriter, r *http.Request) {
	var data PassSignupBody
	logger := RequestLogger(r)

	// Decode the JSON request body into SignupData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Password != data.Password2 {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserByEmail(r.Context(), data.Email, h.pool)

	if err != nil && err != pgx.ErrNoRows {
		logger.Error("Error during signup", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user.ID != 0 {
		// User already exists with this email
		http.Error(w, "Invalid email or password", http.StatusBadRequest)
		return
	}

	// Add any other signup validation logic here
	newUser, err := auth.CreateUser(r.Context(), data.Email, data.Password, h.pool)

	if err != nil {
		slog.Error("Error during signup", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessData := auth.SessionData{
		UserId: newUser.ID,
		Role:   "user",
	}

	token, err := h.jwtManager.EncodeAccessToken(sessData)
	refreshToken, refreshErr := h.jwtManager.EncodeRefreshToken(sessData)

	if errors.Join(err, refreshErr) != nil {
		logger.Error("Error encoding JWT tokens", slog.Any("err", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := LoginJwtResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Error encoding response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) LoginJWT(w http.ResponseWriter, r *http.Request) {
	var data PassLoginBody
	logger := RequestLogger(r)

	// Decode the JSON request body into LoginData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserByEmail(r.Context(), data.Email, h.pool)

	if err != nil {
		utils.ErrorJSON(w, AuthErrorResponse{
			Message: "Invalid email or password",
			Status:  401,
		}, 401)
		return
	}

	err = auth.CheckPasswordHash(data.Password, *user.PasswordHash)

	if err != nil {
		utils.ErrorJSON(w, AuthErrorResponse{
			Message: "Invalid email or password",
			Status:  401,
		}, 401)
		return
	}

	sessData := auth.SessionData{
		UserId: user.ID,
		Role:   user.Role,
	}

	tok, err := h.jwtManager.EncodeAccessToken(sessData)
	refreshTok, refreshErr := h.jwtManager.EncodeRefreshToken(sessData)

	if errors.Join(err, refreshErr) != nil {
		logger.Error("Error encoding JWT tokens", slog.Any("err", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoginJwtResponse{
		AccessToken:  tok,
		RefreshToken: refreshTok,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Error encoding response", slog.Any("err", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

type MeResponse struct {
	Id int `json:"id"`
}

func (h *AuthHandler) GetAuthMe(w http.ResponseWriter, r *http.Request) {

	res := MeResponse{}
	sess, _ := auth.RequestUser(r)

	res.Id = sess.UserId

	err := utils.WriteJSON(w, r, res)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

}
