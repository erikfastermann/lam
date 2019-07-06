package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/erikfastermann/lam/db"
)

func (h Handler) overview(user *db.User, w *response, r *http.Request) (int, string, error) {
	type account struct {
		Color  string
		Banned bool
		Link   string
		DB     db.Account
	}
	type overviewPage struct {
		Username string
		Accounts []account
	}

	db, err := h.db.Accounts()
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("overview: couldn't read accounts from database, %v", err)
	}

	accs := make([]account, 0)
	for _, acc := range db {
		link, _ := URLFromIGN(acc.Region, acc.IGN)

		banned := false
		if acc.Perma {
			banned = true
		} else if acc.Ban.Valid {
			if acc.Ban.Time.Unix()-time.Now().Unix() > 0 {
				banned = true
			}
		} else {
			banned = false
		}

		color := ""
		if banned && !acc.Perma {
			color = "table-warning"
		}
		if acc.Perma || acc.PasswordChanged {
			color = "table-danger"
		}

		accs = append(accs, account{Color: color, Banned: banned, Link: link, DB: *acc})
	}

	data := overviewPage{Username: user.Username, Accounts: accs}
	h.templates.ExecuteTemplate(w, templateOverview, data)
	return http.StatusOK, "", nil
}
