package handler

import (
	"fmt"
	"net/http"

	"github.com/erikfastermann/league-accounts/db"
)

func (h Handler) create(user *db.User, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == http.MethodGet {
		usernames, err := h.db.Usernames()
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Create: Failed querying usernames from database, %v", err)
		}
		acc := db.Account{Region: "euw", User: user.Username}
		data := editPage{Title: "Create new account", Users: usernames, Username: user.Username, Account: acc}
		err = h.templates.ExecuteTemplate(w, "edit.html", data)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("create: method %d not allowed", r.Method)
	}

	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, fmt.Errorf("create: couldn't parse form, %v", err)
	}
	acc, err := accFromForm(r.PostForm)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("create: failed validating form input, %v", err)
	}
	err = h.db.AddAccount(acc)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("create: writing to database failed, %v", err)
	}
	http.Redirect(w, r, "/overview", http.StatusSeeOther)
	return http.StatusOK, nil
}
