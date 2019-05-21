package main

import (
	"fmt"
	"net/http"
	"sort"
	"time"
)

type AccountData struct {
	Color   string
	Banned  bool
	Link    string
	Account AccountDb
}

type AccountsPage struct {
	Username string
	Accounts []AccountData
}

func allAccounts() ([]*AccountDb, error) {
	rows, err := db.Query(`SELECT ID, Region, Tag, Ign, Username, Password, User,
        Leaverbuster, Ban, Perma, PasswordChanged, Pre30, Elo FROM accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accs := make([]*AccountDb, 0)
	for rows.Next() {
		acc := new(AccountDb)
		err := rows.Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.Ign,
			&acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
			&acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accs, nil
}

func accounts(w http.ResponseWriter, r *http.Request) {
	curUser, err := checkAuth(w, r)
	if err != nil {
		return
	}

	accsParsed, err := allAccounts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}

	var accsComputed []AccountData
	var link string

	for _, acc := range accsParsed {
		banned := false
		if acc.Region == "" || acc.Ign == "" {
			link = ""
		} else {
			link = fmt.Sprintf("https://www.leagueofgraphs.com/de/summoner/%s/%s", acc.Region, acc.Ign)
		}

		if acc.Perma {
			banned = true
		} else if acc.Ban.Valid {
			if acc.Ban.Time.Unix()-time.Now().Unix() > 0 {
				banned = true
			}
		} else {
			banned = false
		}

		accsComputed = append(accsComputed, AccountData{Banned: banned, Link: link, Account: *acc})
	}

	sort.SliceStable(accsComputed, func(i, j int) bool {
		return accsComputed[i].Account.Tag < accsComputed[j].Account.Tag
	})

	var accsFinal []AccountData
	for i := 0; i < 3; i++ {
		for _, acc := range accsComputed {
			switch i {
			case 0:
				if !acc.Banned && !acc.Account.PasswordChanged {
					accsFinal = append(accsFinal, acc)
				}
			case 1:
				if acc.Banned && !acc.Account.Perma {
					acc.Color = "table-warning"
					accsFinal = append(accsFinal, acc)
				}
			case 2:
				if acc.Account.Perma || acc.Account.PasswordChanged {
					acc.Color = "table-danger"
					accsFinal = append(accsFinal, acc)
				}
			}
		}
	}

	data := AccountsPage{Username: curUser.Username, Accounts: accsFinal}
	templates.ExecuteTemplate(w, "accounts.html", data)
}
