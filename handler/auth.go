package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/erikfastermann/lam/db"
	"golang.org/x/crypto/bcrypt"
)

func (h Handler) login(user *db.User, w *response, r *http.Request) (int, string, error) {
	if user != nil {
		return http.StatusSeeOther, routeOverview, nil
	}

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, templateLogin, nil)
		return http.StatusUnauthorized, "", nil
	}

	username := r.FormValue("username")
	passwordHash := r.FormValue("password")
	user, err := h.db.User(username)
	if err != nil {
		return http.StatusUnauthorized, routeLogin, fmt.Errorf("couldn't find user (username: %s) in database, %v", username, err)
	}

	byteHash := []byte(passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), byteHash)
	if err != nil {
		return http.StatusUnauthorized, routeLogin, fmt.Errorf("username: %s, %v", username, err)
	}

	randBytes := make([]byte, 24)
	_, err = rand.Read(randBytes)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed generating random bytes, %v", err)
	}
	token := base64.URLEncoding.EncodeToString(randBytes)
	err = h.db.EditToken(user.ID, token)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("couldn't edit token for username: %s, %v", username, err)
	}

	w.cookie = &http.Cookie{
		Name:  "session_token",
		Value: token,
	}
	return http.StatusNoContent, routeOverview, nil
}

func (h Handler) logout(user *db.User, w *response, r *http.Request) (int, string, error) {
	err := h.db.EditToken(user.ID, "")
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("couldn't reset token for username: %s, %v", user.Username, err)
	}
	return http.StatusNoContent, routeLogin, nil
}
