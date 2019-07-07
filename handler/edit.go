package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/erikfastermann/lam/db"
)

func (h Handler) edit(user *db.User, w *response, r *http.Request) (int, string, error) {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("couldn't parse id %s", r.URL.Path[1:])
	}

	if r.Method == http.MethodGet {
		acc, err := h.db.Account(id)
		if err != nil {
			return http.StatusBadRequest, "", fmt.Errorf("couldn't get account with id %d from database, %v", id, err)
		}
		usernames, err := h.db.Usernames()
		if err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("couldn't query usernames from database, %v", err)
		}
		title := fmt.Sprintf("Edit: %s", strconv.Quote(acc.IGN))
		data := editPage{Title: title, Users: usernames, Username: user.Username, Account: *acc}
		h.templates.ExecuteTemplate(w, templateEdit, data)
		return http.StatusOK, "", nil
	}

	acc, err := accFromForm(r)
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("failed validating form input, %v", err)
	}
	err = h.db.EditAccount(id, acc)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("writing account with id %d to database failed, %v", id, err)
	}
	return http.StatusNoContent, routeOverview, nil
}
