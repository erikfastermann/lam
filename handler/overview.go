package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/erikfastermann/lam/db"
)

func (h *Handler) overview(username string, w http.ResponseWriter, r *http.Request) error {
	type account struct {
		Color  string
		Banned bool
		Link   string
		db.Account
	}
	type overviewPage struct {
		Username string
		Accounts []account
	}

	db, err := h.DB.Accounts()
	if err != nil {
		return fmt.Errorf("couldn't read accounts from database, %v", err)
	}

	accs := make([]account, 0)
	for _, acc := range db {
		banned := false
		if acc.Perma || (acc.Ban.Valid && acc.Ban.Time.After(time.Now())) {
			banned = true
		}
		color := ""
		if banned {
			color = "table-warning"
		}
		if acc.Perma || acc.PasswordChanged {
			color = "table-danger"
		}
		accs = append(accs, account{color, banned, LeagueOfGraphsURL(acc.Region, acc.IGN), *acc})
	}

	data := overviewPage{Username: username, Accounts: accs}
	return h.Templates.ExecuteTemplate(w, templateOverview, data)
}
