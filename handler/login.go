package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (h handler) login(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "login.html", nil)
		return nil
	}

	if r.Method != http.MethodPost {
		return statusError{http.StatusMethodNotAllowed, fmt.Errorf("login: method %d not allowed", r.Method)}
	}

	username := r.FormValue("username")
	passwordHash := r.FormValue("password")
	user, err := h.db.User(username)
	if err != nil {
		return statusError{http.StatusUnauthorized, fmt.Errorf("login: couldn't find user (username: %s) in database, %v", username, err)}
	}

	byteHash := []byte(passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), byteHash)
	if err != nil {
		return statusError{http.StatusUnauthorized, fmt.Errorf("login: username: %s, %v", username, err)}
	}

	randBytes := make([]byte, 24)
	_, err = rand.Read(randBytes)
	if err != nil {
		return statusError{http.StatusInternalServerError, fmt.Errorf("login: failed generating random bytes, %v", err)}
	}
	token := base64.URLEncoding.EncodeToString(randBytes)
	err = h.db.EditToken(user.ID, token)
	if err != nil {
		return statusError{http.StatusInternalServerError, fmt.Errorf("login: couldn't edit token for username: %s, %v", username, err)}
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: token,
	})
	http.Redirect(w, r, "/overview", http.StatusSeeOther)
	return nil
}
