package handler

import (
	"fmt"
	"net/http"

	"github.com/erikfastermann/lam/db"
)

func (h Handler) create(user *db.User, w *response, r *http.Request) (int, string, error) {
	if r.Method == http.MethodGet {
		usernames, err := h.db.Usernames()
		if err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("create: failed querying usernames from database, %v", err)
		}
		acc := db.Account{Region: "euw", User: user.Username}
		data := editPage{Title: "Create new account", Users: usernames, Username: user.Username, Account: acc}
		h.templates.ExecuteTemplate(w, templateEdit, data)
		return http.StatusOK, "", nil
	}

	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("create: couldn't parse form, %v", err)
	}
	acc, err := accFromForm(r.PostForm)
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("create: failed validating form input, %v", err)
	}
	err = h.db.AddAccount(acc)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("create: writing to database failed, %v", err)
	}
	return http.StatusCreated, routeOverview, nil
}
