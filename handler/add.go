package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/erikfastermann/lam/db"
)

func (h Handler) add(ctx context.Context, user *db.User, w *response, r *http.Request) (int, string, error) {
	if r.Method == http.MethodGet {
		usernames, err := h.db.Usernames(ctx)
		if err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed querying usernames from database, %v", err)
		}
		acc := db.Account{Region: "euw", User: user.Username}
		data := editPage{Title: "Add new account", Users: usernames, Username: user.Username, Account: acc}
		h.templates.ExecuteTemplate(w, templateEdit, data)
		return http.StatusOK, "", nil
	}

	acc, err := accFromForm(r)
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("failed validating form input, %v", err)
	}
	err = h.db.AddAccount(ctx, acc)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("writing to database failed, %v", err)
	}
	return http.StatusCreated, routeOverview, nil
}
