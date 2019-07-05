package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/erikfastermann/lam/db"
)

type Handler struct {
	db        *db.DB
	templates *template.Template
}

func New(db *db.DB, templates *template.Template) *Handler {
	return &Handler{db: db, templates: templates}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	addr := r.RemoteAddr
	var username string
	var status int
	var errAuth error
	var errRoute error
	user, errAuth := h.checkAuth(r)
	if errAuth != nil {
		username = errAuth.Error()
		status, errRoute = h.router(nil, w, r)
	} else {
		username = fmt.Sprintf("%d:%s", user.ID, user.Username)
		status, errRoute = h.router(user, w, r)
	}
	if errRoute != nil || status == http.StatusNotFound {
		fmt.Fprintf(w, "%d - %s", status, http.StatusText(status))
	}
	log.Printf("%s|%s|%s|%d|%v", addr, url, username, status, errRoute)
}

func (h Handler) router(user *db.User, w http.ResponseWriter, r *http.Request) (int, error) {
	var base string
	base, r.URL.Path = splitURL(r.URL.Path)
	if user == nil {
		return h.login(w, r)
	}

	if base == "edit" {
		return h.edit(user, w, r)
	}
	if base == "create" && r.URL.Path == "/" {
		return h.create(user, w, r)
	}

	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, fmt.Errorf("router: method %s is not allowed", r.Method)
	}
	if base == "remove" {
		return h.remove(user, w, r)
	}

	if r.URL.Path != "/" {
		return http.StatusNotFound, nil
	}
	switch base {
	case "":
		return h.overview(user, w, r)
	case "logout":
		return h.logout(user, w, r)
	}

	return http.StatusNotFound, nil
}

func splitURL(url string) (string, string) {
	url = path.Clean(url)
	split := strings.Split(url[1:], "/")
	if len(split) == 1 {
		return split[0], "/"
	}
	return split[0], "/" + strings.Join(split[1:], "/")
}

func (h Handler) checkAuth(r *http.Request) (*db.User, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	token := c.Value
	user, err := h.db.UserByToken(token)
	if err != nil {
		return nil, fmt.Errorf("auth: Token %s not found, %v", token, err)
	}
	return user, nil
}
