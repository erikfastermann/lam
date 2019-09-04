package handler

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/erikfastermann/httpwrap"
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

type route struct {
	remain  bool
	methods []handlerMethod
	auth    bool
}

type handlerMethod struct {
	method string
	hf     handlerFunc
}

type handlerFunc func(context.Context, *db.User, http.ResponseWriter, *http.Request) error

type Handler struct {
	DB        db.DB
	Templates *template.Template
	routes    map[string]route
	once      sync.Once
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	h.once.Do(h.buildRoutes)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	header := w.Header()
	header.Add("Referrer-Policy", "no-referrer")
	header.Add("X-Frame-Options", "DENY")
	header.Add("X-Content-Type-Options", "nosniff")
	if r.URL.Scheme == "https" {
		header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}

	var base string
	base, r.URL.Path = splitURL(r.URL.Path)

	rt, ok := h.routes[base]
	if !ok || rt.remain != (r.URL.Path != "/") {
		return httpwrap.Error{StatusCode: http.StatusNotFound, Err: errors.New("unknown route")}
	}

	user, err := h.checkAuth(ctx, r)
	if err != nil && rt.auth {
		http.Redirect(w, r, routeLogin, http.StatusSeeOther)
		return err
	}

	for _, m := range rt.methods {
		if m.method == r.Method {
			return m.hf(ctx, user, w, r)
		}
	}
	return httpwrap.Error{StatusCode: http.StatusMethodNotAllowed, Err: errors.New("bad method")}
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
