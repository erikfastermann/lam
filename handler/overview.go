package handler

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/erikfastermann/league-accounts/db"
)

type account struct {
	Color  string
	Banned bool
	Link   string
	DB     db.Account
}

type accountsPage struct {
	Username string
	Accounts []account
}

func (h handler) overview(username string, w http.ResponseWriter, r *http.Request) error {
	accsDb, err := h.db.Accounts()
	if err != nil {
		return statusError{http.StatusInternalServerError, fmt.Errorf("overview: couldn't read accounts from database, %v", err)}
	}

	accs := make([]account, 0)
	for _, acc := range accsDb {
		banned := false
		link, _ := URLFromIGN(acc.Region, acc.IGN)
		if acc.Perma {
			banned = true
		} else if acc.Ban.Valid {
			if acc.Ban.Time.Unix()-time.Now().Unix() > 0 {
				banned = true
			}
		} else {
			banned = false
		}
		accs = append(accs, account{Banned: banned, Link: link, DB: *acc})
	}

	sort.SliceStable(accs, func(i, j int) bool {
		return accs[i].DB.Tag < accs[j].DB.Tag
	})

	accsSortedByBan := make([]account, 0)
	for i := 0; i < 3; i++ {
		for _, acc := range accs {
			if i == 0 && (!acc.Banned && !acc.DB.PasswordChanged) {
				accsSortedByBan = append(accsSortedByBan, acc)
			}
			if i == 1 && (acc.Banned && !acc.DB.Perma) {
				acc.Color = "table-warning"
				accsSortedByBan = append(accsSortedByBan, acc)
			}
			if i == 2 && (acc.DB.Perma || acc.DB.PasswordChanged) {
				acc.Color = "table-danger"
				accsSortedByBan = append(accsSortedByBan, acc)
			}
		}
	}

	data := accountsPage{Username: username, Accounts: accsSortedByBan}
	h.templates.ExecuteTemplate(w, "overview.html", data)
	return nil
}
