package auth

import (
	"net/http"

	"github.com/maybemaby/smolauth"
)

type AuthHandler struct {
	manager *smolauth.AuthManager
}

func NewAuthHandler(manager *smolauth.AuthManager) *AuthHandler {
	return &AuthHandler{
		manager: manager,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	
}
