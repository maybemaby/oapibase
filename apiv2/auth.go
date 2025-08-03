package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/maybemaby/oapibase/apiv2/utils"
	"github.com/maybemaby/smolauth"
	"github.com/oapi-codegen/runtime/types"
)

type AuthHandler struct {
	authManager *smolauth.AuthManager
}

type PassLoginBody struct {
	Email    string `json:"email" example:"email@site.com"`
	Password string `json:"password"`
}

// PostAuthLogin implements gen.ServerInterface.
func (h *AuthHandler) PostAuthLogin(w http.ResponseWriter, r *http.Request) {

	data := PassLoginBody{}

	err := utils.ReadJSON(r, &data)

	if err != nil {
		http.Error(w, "Bad JSON body", http.StatusBadRequest)
		return
	}

	id, err := h.authManager.CheckPassword(string(data.Email), data.Password)

	if err != nil {
		http.Error(w, "Invalid Password or Email", http.StatusUnauthorized)
		return
	}

	err = h.authManager.Login(r, smolauth.SessionData{
		UserId: id,
	})

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostAuthLogout implements gen.ServerInterface.
func (h *AuthHandler) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	err := h.authManager.Logout(r)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type PassSignupBody struct {
	Email     string `json:"email" example:"email@site.com"`
	Password  string `json:"password" minLength:"8"`
	Password2 string `json:"password2"`
}

// PostAuthSignup implements gen.ServerInterface.
func (h *AuthHandler) PostAuthSignup(w http.ResponseWriter, r *http.Request) {

	logger := RequestLogger(r)
	data := PassSignupBody{}

	err := utils.ReadJSON(r, &data)

	if err != nil {

		if errors.Is(err, types.ErrValidationEmail) {
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			return
		}

		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if len(data.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	if data.Password != data.Password2 {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	id, err := h.authManager.PasswordSignup(string(data.Email), data.Password)

	if err != nil {
		logger.Error("Error signing up user", slog.String("err", err.Error()))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = h.authManager.Login(r, smolauth.SessionData{
		UserId: id,
	})

	if err != nil {
		logger.Error("Error logging in user", slog.String("err", err.Error()))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type MeResponse struct {
	Id int `json:"id"`
}

func (h *AuthHandler) GetAuthMe(w http.ResponseWriter, r *http.Request) {

	mw := smolauth.RequireAuthMiddleware(h.authManager)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		res := MeResponse{}
		sess, _ := h.authManager.GetSession(r)

		res.Id = sess.UserId

		err := utils.WriteJSON(w, r, res)

		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

	})).ServeHTTP(w, r)

}
