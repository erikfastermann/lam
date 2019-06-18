package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/erikfastermann/league-accounts/db"
	"golang.org/x/crypto/bcrypt"
)

func (h Handler) login(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == http.MethodGet {
		err := h.templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("login: method %d not allowed", r.Method)
	}

	username := r.FormValue("username")
	passwordHash := r.FormValue("password")
	user, err := h.db.User(username)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("login: couldn't find user (username: %s) in database, %v", username, err)
	}

	byteHash := []byte(passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), byteHash)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("login: username: %s, %v", username, err)
	}

	randBytes := make([]byte, 24)
	_, err = rand.Read(randBytes)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("login: failed generating random bytes, %v", err)
	}
	token := base64.URLEncoding.EncodeToString(randBytes)
	err = h.db.EditToken(user.ID, token)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("login: couldn't edit token for username: %s, %v", username, err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: token,
	})
	http.Redirect(w, r, "/overview", http.StatusSeeOther)
	return http.StatusOK, nil
}

func (h Handler) logout(user *db.User, w http.ResponseWriter, r *http.Request) (int, error) {
	err := h.db.EditToken(user.ID, "")
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("logout: couldn't reset token for username: %s, %v", user.Username, err)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return http.StatusOK, nil
}