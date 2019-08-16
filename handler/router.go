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
	"sync"
	"time"

	"github.com/erikfastermann/lam/db"
)

const (
	routeLogin    = "/login"
	routeLogout   = "/logout"
	routeOverview = "/"
	routeEdit     = "/edit"
	routeAdd      = "/add"
	routeRemove   = "/remove"
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

type route struct {
	remain  bool
	methods []handlerMethod
	auth    bool
}

type handlerMethod struct {
	method string
	hf     handlerFunc
}

type handlerFunc func(context.Context, *db.User, *response, *http.Request) (int, string, error)

type Handler struct {
	DB        db.DB
	Templates *template.Template
	HTTPS     bool
	Logger    *log.Logger
	routes    map[string]route
	once      sync.Once
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.once.Do(h.buildRoutes)

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

func (h *Handler) buildRoutes() {
	h.routes = map[string]route{
		routeLogin: {
			false,
			[]handlerMethod{
				{http.MethodGet, h.login},
				{http.MethodPost, h.login},
			},
			false,
		},
		routeLogout: {
			false,
			[]handlerMethod{
				{http.MethodGet, h.logout},
			},
			true,
		},
		routeOverview: {
			false,
			[]handlerMethod{
				{http.MethodGet, h.overview},
			},
			true,
		},
		routeAdd: {
			false,
			[]handlerMethod{
				{http.MethodGet, h.add},
				{http.MethodPost, h.add},
			},
			true,
		},
		routeEdit: {
			true,
			[]handlerMethod{
				{http.MethodGet, h.edit},
				{http.MethodPost, h.edit},
			},
			true,
		},
		routeRemove: {
			true,
			[]handlerMethod{
				{http.MethodGet, h.remove},
			},
			true,
		},
	}
}

func (h *Handler) handleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (status int, handlerErr error, user *db.User, authErr error) {
	user, authErr = h.checkAuth(ctx, r)
	var redirect string
	resp := new(response)
	status, redirect, handlerErr = h.router(ctx, user, authErr, resp, r)
	if resp.cookie != nil {
		http.SetCookie(w, resp.cookie)
	}

	header := w.Header()
	header.Add("Referrer-Policy", "no-referrer")
	header.Add("X-Frame-Options", "DENY")
	header.Add("X-Content-Type-Options", "nosniff")
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

func (h *Handler) router(ctx context.Context, user *db.User, authErr error, w *response, r *http.Request) (int, string, error) {
	var base string
	base, r.URL.Path = splitURL(r.URL.Path)

	rt, ok := h.routes[base]
	if !ok || rt.remain != (r.URL.Path != "/") {
		return http.StatusNotFound, "", nil
	}

	if rt.auth && authErr != nil {
		return http.StatusUnauthorized, routeLogin, nil
	}

	for _, m := range rt.methods {
		if m.method == r.Method {
			return m.hf(ctx, user, w, r)
		}
	}
	return http.StatusMethodNotAllowed, "", nil
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

func (h *Handler) checkAuth(ctx context.Context, r *http.Request) (*db.User, error) {
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
