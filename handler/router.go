package handler

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/erikfastermann/lam/db"
)

const (
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
	DB        *db.DB
	Templates *template.Template
	HTTPS     bool
	Logger    *log.Logger
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		panic("db is nil")
	}
	if h.Templates == nil {
		panic("templates is nil")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	url := r.URL.Path
	addr := r.RemoteAddr
	var username string
	status, handlerErr, user, authErr := h.handleRequest(ctx, w, r)
	if authErr != nil {
		username = authErr.Error()
	} else {
		username = fmt.Sprintf("%d:%s", user.ID, user.Username)
	}
	if h.Logger != nil {
		h.Logger.Printf("%s|%s %s|%s|%d - %s|%v", addr, r.Method, url, username, status, http.StatusText(status), handlerErr)
	}
}

func (h Handler) handleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (status int, handlerErr error, user *db.User, authErr error) {
	user, authErr = h.checkAuth(ctx, r)
	var redirect string
	resp := new(response)
	status, redirect, handlerErr = h.router(ctx, user, authErr, resp, r)
	if resp.cookie != nil {
		http.SetCookie(w, resp.cookie)
	}

	header := w.Header()
	if h.HTTPS {
		header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}

	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	if handlerErr != nil || status == http.StatusNotFound {
		header.Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, "%d - %s", status, http.StatusText(status))
		return
	}

	header.Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
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
	routeAdd      = "/add"
	routeRemove   = "/remove"
)

func (h Handler) router(ctx context.Context, user *db.User, authErr error, w *response, r *http.Request) (int, string, error) {
	routes := []struct {
		f    func(context.Context, *db.User, *response, *http.Request) (int, string, error)
		base string
		id   bool
		post bool
		auth bool
	}{
		{f: h.login, base: routeLogin, id: false, post: true, auth: false},
		{f: h.logout, base: routeLogout, id: false, post: false, auth: true},
		{f: h.overview, base: routeOverview, id: false, post: false, auth: true},
		{f: h.edit, base: routeEdit, id: true, post: true, auth: true},
		{f: h.add, base: routeAdd, id: false, post: true, auth: true},
		{f: h.remove, base: routeRemove, id: true, post: false, auth: true},
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
		return i.f(ctx, user, w, r)
	}
	return http.StatusNotFound, "", nil
}

func splitURL(url string) (string, string) {
	url = path.Clean(url)
	split := strings.SplitN(url[1:], "/", 2)
	split[0] = "/" + split[0]
	if len(split) == 1 {
		return split[0], "/"
	}
	return split[0], "/" + split[1]
}

func (h Handler) checkAuth(ctx context.Context, r *http.Request) (*db.User, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	token := c.Value
	user, err := h.DB.UserByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("auth: Token %s not found, %v", token, err)
	}
	return user, nil
}
