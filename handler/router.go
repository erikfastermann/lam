package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/erikfastermann/lam/db"
)

var (
	templateLogin    = "login.html"
	templateOverview = "overview.html"
	templateEdit     = "edit.html"
)

type response struct {
	buf    bytes.Buffer
	cookie *http.Cookie
}

func (r *response) Write(p []byte) (int, error) {
	return r.buf.Write(p)
}

type Handler struct {
	db        *db.DB
	templates *template.Template
	logger    *log.Logger
}

func New(db *db.DB, templates *template.Template, logger *log.Logger) *Handler {
	return &Handler{db: db, templates: templates, logger: logger}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	addr := r.RemoteAddr
	var username string
	status, handlerErr, user, authErr := h.handleRequest(w, r)
	if authErr != nil {
		username = authErr.Error()
	} else {
		username = fmt.Sprintf("%d:%s", user.ID, user.Username)
	}
	h.logger.Printf("%s|%s %s|%s|%d - %s|%v", addr, r.Method, url, username, status, http.StatusText(status), handlerErr)
}

func (h Handler) handleRequest(w http.ResponseWriter, r *http.Request) (status int, handlerErr error, user *db.User, authErr error) {
	user, authErr = h.checkAuth(r)
	var redirect string
	resp := new(response)
	status, redirect, handlerErr = h.router(user, authErr, resp, r)
	if resp.cookie != nil {
		http.SetCookie(w, resp.cookie)
	}
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}
	w.WriteHeader(status)
	if handlerErr != nil || status == http.StatusNotFound {
		fmt.Fprintf(w, "%d - %s", status, http.StatusText(status))
		return
	}
	_, err := resp.buf.WriteTo(w)
	if err != nil {
		handlerErr = fmt.Errorf("router: failed writing template, %v", err)
		return
	}
	return
}

const (
	routeLogin    = "/login"
	routeLogout   = "/logout"
	routeOverview = "/"
	routeEdit     = "/edit"
	routeCreate   = "/create"
	routeRemove   = "/remove"
)

func (h Handler) router(user *db.User, authErr error, w *response, r *http.Request) (int, string, error) {
	routes := []struct {
		f    func(*db.User, *response, *http.Request) (int, string, error)
		base string
		id   bool
		post bool
		auth bool
		wrap bool
	}{
		{f: h.login, base: routeLogin[1:], id: false, post: true, auth: false},
		{f: h.logout, base: routeLogout[1:], id: false, post: false, auth: true},
		{f: h.overview, base: routeOverview[1:], id: false, post: false, auth: true},
		{f: h.edit, base: routeEdit[1:], id: true, post: true, auth: true},
		{f: h.create, base: routeCreate[1:], id: false, post: true, auth: true},
		{f: h.remove, base: routeRemove[1:], id: true, post: false, auth: true},
	}

	var base string
	base, r.URL.Path = splitURL(r.URL.Path)
	for _, i := range routes {
		if i.base != base || (!i.id && r.URL.Path != "/") {
			continue
		}
		if (!i.post && r.Method == http.MethodPost) || !(r.Method == http.MethodGet || r.Method == http.MethodPost) {
			return http.StatusMethodNotAllowed, "", fmt.Errorf("router: method %s is not allowed", r.Method)
		}
		if i.auth && authErr != nil {
			return http.StatusUnauthorized, routeLogin, nil
		}
		return i.f(user, w, r)
	}
	return http.StatusNotFound, "", nil
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
