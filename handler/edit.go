package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erikfastermann/httpwrap"
	"github.com/erikfastermann/lam/db"
)

func (h *Handler) edit(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return httpwrap.Error{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("couldn't parse id %s", r.URL.Path[1:]),
		}
	}

	if r.Method == http.MethodGet {
		acc, err := h.DB.Account(ctx, id)
		if err != nil {
			return httpwrap.Error{
				StatusCode: http.StatusBadRequest,
				Err:        fmt.Errorf("couldn't get account with id %d from database, %v", id, err),
			}
		}
		usernames, err := h.DB.Usernames(ctx)
		if err != nil {
			return fmt.Errorf("couldn't query usernames from database, %v", err)
		}
		title := fmt.Sprintf("Edit: %s", strconv.Quote(acc.IGN))
		data := editPage{Title: title, Users: usernames, Username: user.Username, Account: *acc}
		return h.Templates.ExecuteTemplate(w, templateEdit, data)
	}

	acc, err := accFromForm(r)
	if err != nil {
		return httpwrap.Error{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("failed validating form input, %v", err),
		}
	}
	err = h.DB.EditAccount(ctx, id, acc)
	if err != nil {
		if err == sql.ErrNoRows {
			return httpwrap.Error{
				StatusCode: http.StatusBadRequest,
				Err:        fmt.Errorf("couldn't find account with id %d", id),
			}
		}
		return fmt.Errorf("writing account with id %d failed, %v", id, err)
	}
	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}
