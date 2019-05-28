package handler

import (
	"fmt"
	"net/http"
	"strconv"
)

func (h handler) edit(username string, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		return statusError{http.StatusBadRequest, fmt.Errorf("edit: couldn't parse id, %v", err)}
	}

	if r.Method == http.MethodGet {
		acc, err := h.db.Account(id)
		if err != nil {
			return statusError{http.StatusBadRequest, fmt.Errorf("edit: couldn't get account with id %d from database, %v", id, err)}
		}
		usernames, err := h.db.Usernames()
		if err != nil {
			return statusError{http.StatusInternalServerError, fmt.Errorf("edit: couldn't query usernames from database, %v", err)}
		}
		title := fmt.Sprintf("Edit: %s", acc.IGN)
		data := editPage{Title: title, Users: usernames, Username: username, Account: *acc}
		h.templates.ExecuteTemplate(w, "edit.html", data)
		return nil
	}

	if r.Method != http.MethodPost {
		return statusError{http.StatusMethodNotAllowed, fmt.Errorf("edit: method %d not allowed", r.Method)}
	}

	if err := r.ParseForm(); err != nil {
		return statusError{http.StatusBadRequest, fmt.Errorf("edit: couldn't parse form, %v", err)}
	}
	acc, err := accFromForm(r.PostForm)
	if err != nil {
		return statusError{http.StatusBadRequest, fmt.Errorf("edit: failed validating form input, %v", err)}
	}
	err = h.db.EditAccount(id, acc)

	if err != nil {
		return statusError{http.StatusInternalServerError, fmt.Errorf("edit: writing account with id %d to database failed, %v", id, err)}
	}
	http.Redirect(w, r, "/overview", http.StatusSeeOther)
	return nil
}
