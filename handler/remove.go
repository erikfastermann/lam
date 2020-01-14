package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

func (h *Handler) remove(_ string, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return badRequestf("couldn't parse id %s", r.URL.Path[1:])
	}

	if err := h.DB.RemoveAccount(id); err != nil {
		if err == sql.ErrNoRows {
			return badRequestf("couldn't find account with id %d", id)
		}
		return fmt.Errorf("couldn't remove account with id %d, %v", id, err)
	}

	http.Redirect(w, r, routeOverview, http.StatusSeeOther)
	return nil
}
