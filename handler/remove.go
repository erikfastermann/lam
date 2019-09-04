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

func (h *Handler) remove(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return httpwrap.Error{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("couldn't parse id %s", r.URL.Path[1:]),
		}
	}
	err = h.DB.RemoveAccount(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return httpwrap.Error{
				StatusCode: http.StatusBadRequest,
				Err:        fmt.Errorf("couldn't find account with id %d", id),
			}
		}
		return fmt.Errorf("couldn't remove account with id %d, %v", id, err)
	}
	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}
