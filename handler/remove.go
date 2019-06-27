package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/erikfastermann/lam/db"
)

func (h Handler) remove(user *db.User, w http.ResponseWriter, r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("remove: couldn't parse id %s", r.URL.Path[1:])
	}
	_, err = h.db.Account(id)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("remove: couldn't find to be removed account with id %d from database, %v", id, err)
	}
	err = h.db.RemoveAccount(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("remove: couldn't remove account with id %d from database, %v", id, err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return http.StatusNoContent, nil
}
