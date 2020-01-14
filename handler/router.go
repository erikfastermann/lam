package handler

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
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

type User struct {
	Username, Password, Token string
}

var errBadMethod = httpwrap.Error{
	StatusCode: http.StatusMethodNotAllowed,
	Err:        errors.New("bad method"),
}

type route struct {
	remain  bool
	methods []string
	hf      handlerFunc
}

type handlerFunc func(ctx context.Context, username string, w http.ResponseWriter, r *http.Request) error

type Handler struct {
	DB *db.DB

	mu    sync.RWMutex
	Users []*User

	Templates *template.Template

	once            sync.Once
	logger          *log.Logger
	protectedRoutes map[string]route
}

func (h *Handler) ServeHTTPWithErr(w http.ResponseWriter, r *http.Request) error {
	err := h.serve(w, r)
	if httpwrap.IsErrorInternal(err) {
		h.logger.Print(err)
	}
	return err
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request) error {
	h.once.Do(h.setup)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	header := w.Header()
	header.Add("Referrer-Policy", "no-referrer")
	header.Add("X-Frame-Options", "DENY")
	header.Add("X-Content-Type-Options", "nosniff")
	if r.URL.Scheme == "https" {
		header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}

	username, err := h.checkAuth(r)
	if path.Clean(r.URL.Path) == routeLogin {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			return errBadMethod
		}
		return h.login(ctx, username, w, r)
	}
	if err != nil {
		http.Redirect(w, r, routeLogin, http.StatusSeeOther)
		return err
	}

	fn, err := h.router(r)
	if err != nil {
		return err
	}
	return fn(ctx, username, w, r)
}

func (h *Handler) router(r *http.Request) (handlerFunc, error) {
	var base string
	base, r.URL.Path = splitURL(r.URL.Path)

	rt, ok := h.protectedRoutes[base]
	if !ok || rt.remain != (r.URL.Path != "/") {
		return nil, httpwrap.Error{
			StatusCode: http.StatusNotFound,
			Err:        errors.New("unknown route"),
		}
	}

	for _, method := range rt.methods {
		if method == r.Method {
			return rt.hf, nil
		}
	}
	return nil, errBadMethod
}

func (h *Handler) setup() {
	h.logger = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	h.protectedRoutes = map[string]route{
		routeLogout: {
			false,
			[]string{http.MethodGet},
			h.logout,
		},
		routeOverview: {
			false,
			[]string{http.MethodGet},
			h.overview,
		},
		routeAdd: {
			false,
			[]string{http.MethodGet, http.MethodPost},
			h.add,
		},
		routeEdit: {
			true,
			[]string{http.MethodGet, http.MethodPost},
			h.edit,
		},
		routeRemove: {
			true,
			[]string{http.MethodGet},
			h.remove,
		},
	}
}

func (h *Handler) usernames() []string {
	usernames := make([]string, 0)
	h.mu.RLock()
	for _, u := range h.Users {
		usernames = append(usernames, u.Username)
	}
	h.mu.RUnlock()
	return usernames
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
