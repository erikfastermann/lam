package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/erikfastermann/httpwrap"
	"github.com/erikfastermann/lam/db"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) checkAuth(ctx context.Context, r *http.Request) (*db.User, bool, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil, false, fmt.Errorf("auth: failed reading cookie, %v", err)
	}

	token := c.Value
	user, err := h.DB.UserByToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, fmt.Errorf("auth: token %s not found in database", token)
		}
		return nil, true, err
	}
	return user, false, nil
}

func (h *Handler) login(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) error {
	errUnauthorized := httpwrap.Error{
		StatusCode: http.StatusUnauthorized,
		Err:        errors.New("unauthorized"),
	}

	if user != nil {
		http.Redirect(w, r, routeOverview, http.StatusSeeOther)
		return nil
	}

	if r.Method == http.MethodGet {
		if err := h.Templates.ExecuteTemplate(w, templateLogin, nil); err != nil {
			return err
		}
		return errUnauthorized
	}

	username := r.FormValue("username")
	passwordHash := r.FormValue("password")
	user, err := h.DB.User(ctx, username)
	if err != nil {
		http.Redirect(w, r, routeLogin, http.StatusSeeOther)
		errUnauthorized.Err = fmt.Errorf("couldn't find user (username: %s) in database, %v", username, err)
		return errUnauthorized
	}

	byteHash := []byte(passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), byteHash)
	if err != nil {
		http.Redirect(w, r, routeLogin, http.StatusSeeOther)
		errUnauthorized.Err = fmt.Errorf("username: %s, %v", username, err)
		return errUnauthorized
	}

	randBytes := make([]byte, 24)
	_, err = rand.Read(randBytes)
	if err != nil {
		return fmt.Errorf("failed generating random bytes, %v", err)
	}
	token := base64.URLEncoding.EncodeToString(randBytes)
	err = h.DB.EditToken(ctx, user.ID, token)
	if err != nil {
		return fmt.Errorf("couldn't edit token for username: %s, %v", username, err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}

func (h *Handler) logout(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) error {
	err := h.DB.EditToken(ctx, user.ID, "")
	if err != nil {
		return fmt.Errorf("couldn't reset token for username: %s, %v", user.Username, err)
	}
	http.Redirect(w, r, routeLogin, http.StatusSeeOther)
	return nil
}
