package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/maybemaby/oapibase/api/gen"
	"github.com/maybemaby/oapibase/api/utils"
	"github.com/maybemaby/smolauth"
	"github.com/oapi-codegen/runtime/types"
)

// PostAuthLogin implements gen.ServerInterface.
func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {

	data := gen.PassLoginBody{}

	err := utils.ReadJSON(r, &data)

	if err != nil {
		http.Error(w, "Bad JSON body", http.StatusBadRequest)
		return
	}

	id, err := s.authManager.CheckPassword(string(data.Email), data.Password)

	if err != nil {
		http.Error(w, "Invalid Password or Email", http.StatusUnauthorized)
		return
	}

	err = s.authManager.Login(r, smolauth.SessionData{
		UserId: id,
	})

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostAuthLogout implements gen.ServerInterface.
func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	err := s.authManager.Logout(r)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostAuthSignup implements gen.ServerInterface.
func (s *Server) PostAuthSignup(w http.ResponseWriter, r *http.Request) {

	logger := RequestLogger(r)
	data := gen.PassSignupBody{}

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

	id, err := s.authManager.PasswordSignup(string(data.Email), data.Password)

	if err != nil {
		logger.Error("Error signing up user", slog.String("err", err.Error()))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = s.authManager.Login(r, smolauth.SessionData{
		UserId: id,
	})

	if err != nil {
		logger.Error("Error logging in user", slog.String("err", err.Error()))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetAuthMe(w http.ResponseWriter, r *http.Request) {

	mw := smolauth.RequireAuthMiddleware(s.authManager)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		res := &gen.MeResponse{}
		sess, _ := s.authManager.GetSession(r)

		res.Id = sess.UserId

		err := utils.WriteJSON(w, r, res)

		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

	})).ServeHTTP(w, r)

}
