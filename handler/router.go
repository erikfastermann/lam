package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/erikfastermann/league-accounts/db"
)

type statusError struct {
	code int
	err  error
}

func (e statusError) Error() string {
	return e.err.Error()
}

type handler struct {
	db        *db.Service
	templates *template.Template
}

type loggingHandler func(w http.ResponseWriter, r *http.Request) error
type authHandler func(username string, w http.ResponseWriter, r *http.Request) error

type editPage struct {
	Title    string
	Users    []string
	Username string
	Account  db.Account
}

func withLogging(h loggingHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		statusErr := h(w, r)
		if statusErr != nil {
			if err, ok := statusErr.(statusError); ok {
				log.Println(path, err.code, err.Error())
				if err.code == http.StatusUnauthorized {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				http.Error(w, http.StatusText(err.code), err.code)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Println(path, http.StatusOK)
	}
}

func (h handler) withAuth(ah authHandler) loggingHandler {
	return loggingHandler(func(w http.ResponseWriter, r *http.Request) error {
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				return statusError{http.StatusUnauthorized, err}
			}
			return statusError{http.StatusBadRequest, err}
		}
		token := c.Value
		username, err := h.db.UsernameByToken(token)
		if err != nil {
			return statusError{http.StatusUnauthorized, fmt.Errorf("auth: Token %s not found, %v", token, err)}
		}
		return ah(username, w, r)
	})
}

func New(db *db.Service, templates *template.Template) http.Handler {
	mux := http.NewServeMux()
	h := handler{db, templates}
	mux.Handle("/", withLogging(h.withAuth(h.overview)))
	mux.Handle("/overview", withLogging(h.withAuth(h.overview)))
	mux.Handle("/create", withLogging(h.withAuth(h.create)))
	mux.Handle("/edit", withLogging(h.withAuth(h.edit)))
	mux.Handle("/login", withLogging(h.login))
	return mux
}
