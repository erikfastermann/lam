package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

func (h *Handler) edit(username string, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return badRequestf("couldn't parse id %s", r.URL.Path[1:])
	}

	if r.Method == http.MethodGet {
		acc, err := h.DB.Account(id)
		if err != nil {
			return badRequestf("couldn't get account with id %d from database, %v", id, err)
		}

		title := fmt.Sprintf("Edit: %s", strconv.Quote(acc.IGN))
		data := editPage{Title: title, Users: h.usernames(), Username: username, Account: *acc}
		return h.Templates.ExecuteTemplate(w, templateEdit, data)
	}

	acc, err := accFromForm(r)
	if err != nil {
		return badRequestf("failed validating form input, %v", err)
	}

	if err := h.DB.EditAccount(id, acc); err != nil {
		if err == sql.ErrNoRows {
			badRequestf("couldn't find account with id %d", id)
		}
		return fmt.Errorf("writing account with id %d failed, %v", id, err)
	}

	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}
