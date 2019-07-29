package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erikfastermann/lam/db"
)

func (h Handler) remove(ctx context.Context, user *db.User, w *response, r *http.Request) (int, string, error) {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("couldn't parse id %s", r.URL.Path[1:])
	}
	err = h.DB.RemoveAccount(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return http.StatusBadRequest, "", fmt.Errorf("couldn't find account with id %d", id)
		}
		return http.StatusInternalServerError, "", fmt.Errorf("couldn't remove account with id %d, %v", id, err)
	}
	return http.StatusNoContent, routeOverview, nil
}
