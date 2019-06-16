package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/erikfastermann/league-accounts/db"
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
	status, err := h.router(w, r)
	if err != nil || status == http.StatusNotFound {
		fmt.Fprintf(w, "%d - %s", status, http.StatusText(status))
	}
	msg := fmt.Sprintf("%s - %d", url, status)
	if err != nil {
		msg += fmt.Sprintf(" - %v", err)
	}
	log.Printf(msg)
}

func (h Handler) router(w http.ResponseWriter, r *http.Request) (int, error) {
	var base string
	base, r.URL.Path = splitURL(r.URL.Path)
	username, err := h.checkAuth(r)
	if err != nil {
		return h.login(w, r)
	}
	if base == "edit" {
		return h.edit(username, w, r)
	}
	if r.URL.Path != "/" {
		return http.StatusNotFound, nil
	}
	switch base {
	case "":
		http.Redirect(w, r, "/overview", http.StatusMovedPermanently)
		return http.StatusMovedPermanently, nil
	case "overview":
		return h.overview(username, w, r)
	case "create":
		return h.create(username, w, r)
	case "login":
		return h.login(w, r)
	default:
		return http.StatusNotFound, nil
	}
}

func splitURL(url string) (string, string) {
	url = path.Clean(url)
	split := strings.Split(url[1:], "/")
	if len(split) == 1 {
		return split[0], "/"
	}
	return split[0], "/" + strings.Join(split[1:], "/")
}

func (h Handler) checkAuth(r *http.Request) (string, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	token := c.Value
	username, err := h.db.UsernameByToken(token)
	if err != nil {
		return "", fmt.Errorf("auth: Token %s not found, %v", token, err)
	}
	return username, nil
}
