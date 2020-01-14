package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/erikfastermann/httpwrap"
	"golang.org/x/crypto/bcrypt"
)

const sessToken = "session_token"

func unauthf(format string, a ...interface{}) error {
	return httpwrap.Error{
		StatusCode: http.StatusUnauthorized,
		Err:        fmt.Errorf(format, a...),
	}
}

func (h *Handler) findUsername(username string) (*User, bool) {
	for _, u := range h.Users {
		if u.Username != username {
			continue
		}
		return u, true
	}
	return nil, false
}

func (h *Handler) findUser(token string) (*User, bool) {
	for _, u := range h.Users {
		if u.Token != token {
			continue
		}
		return u, true
	}
	return nil, false
}

func (h *Handler) signIn(username, password string) (token string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	u, ok := h.findUsername(username)
	if !ok {
		return "", unauthf("username %q doesn't exist", username)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", unauthf("username: %q, %v", username, err)
	}

	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	u.Token = base64.URLEncoding.EncodeToString(b)

	return u.Token, nil
}

func (h *Handler) checkAuth(r *http.Request) (string, error) {
	c, err := r.Cookie(sessToken)
	if err != nil {
		return "", err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	token := c.Value
	u, ok := h.findUser(token)
	if !ok {
		return "", unauthf("token %q doesn't exist", token)
	}

	return u.Username, nil
}

func (h *Handler) login(username string, w http.ResponseWriter, r *http.Request) error {
	if username != "" {
		http.Redirect(w, r, routeOverview, http.StatusSeeOther)
		return nil
	}

	if r.Method == http.MethodGet {
		if err := h.Templates.ExecuteTemplate(w, templateLogin, nil); err != nil {
			return err
		}
		return unauthf("unauthorized")
	}

	token, err := h.signIn(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		http.Redirect(w, r, routeLogin, http.StatusSeeOther)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessToken,
		Value:    token,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}

func (h *Handler) logout(username string, w http.ResponseWriter, r *http.Request) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	u, ok := h.findUsername(username)
	if !ok {
		return fmt.Errorf("logout: username %q not found", username)
	}
	u.Token = ""

	http.Redirect(w, r, routeLogin, http.StatusSeeOther)
	return nil
}
