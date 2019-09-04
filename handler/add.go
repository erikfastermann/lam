package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/erikfastermann/httpwrap"
	"github.com/erikfastermann/lam/db"
)

func (h *Handler) add(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		usernames, err := h.DB.Usernames(ctx)
		if err != nil {
			return fmt.Errorf("failed querying usernames from database, %v", err)
		}
		acc := db.Account{Region: "euw", User: user.Username}
		data := editPage{Title: "Add new account", Users: usernames, Username: user.Username, Account: acc}
		return h.Templates.ExecuteTemplate(w, templateEdit, data)
	}

	acc, err := accFromForm(r)
	if err != nil {
		return httpwrap.Error{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("failed validating form input, %v", err),
		}
	}
	err = h.DB.AddAccount(ctx, acc)
	if err != nil {
		return fmt.Errorf("writing to database failed, %v", err)
	}
	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}
