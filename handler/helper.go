package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/erikfastermann/lam/db"
)

type editPage struct {
	Title    string
	Users    []string
	Username string
	Account  db.Account
}

func accFromForm(r *http.Request) (*db.Account, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	formVal := func(key string) string {
		val, ok := r.PostForm[key]
		if !ok {
			return ""
		}
		return val[0]
	}

	acc := new(db.Account)
	acc.Region = formVal("region")
	acc.Tag = formVal("tag")
	acc.IGN = formVal("ign")
	acc.Username = formVal("username")
	acc.Password = formVal("password")
	acc.User = formVal("user")

	leaverbuster := formVal("leaverbuster")
	leaverbusterInt, err := strconv.Atoi(leaverbuster)
	if err != nil {
		return nil, err
	}
	acc.Leaverbuster = leaverbusterInt

	banForm := formVal("ban")
	var ban db.NullTime
	if banForm != "" {
		ban.Time, err = time.ParseInLocation("2006-01-02 15:04", banForm, time.Local)
		if err != nil {
			return nil, err
		}
		ban.Valid = true
	}
	acc.Ban = ban

	toBool := func(str string) (bool, error) {
		if str == "true" {
			return true, nil
		}
		if str == "false" || str == "" {
			return false, nil
		}
		return false, fmt.Errorf("failed converting %s to bool", str)
	}

	perma := formVal("perma")
	acc.Perma, err = toBool(perma)
	if err != nil {
		return nil, fmt.Errorf("form-perma: %v", err)
	}
	pwChanged := formVal("password_changed")
	acc.PasswordChanged, err = toBool(pwChanged)
	if err != nil {
		return nil, fmt.Errorf("form-password_changed: %v", err)
	}
	pre30 := formVal("pre_30")
	acc.Pre30, err = toBool(pre30)
	if err != nil {
		return nil, fmt.Errorf("form-pre_30: %v", err)
	}
	return acc, nil
}

func LeagueOfGraphsURL(region, ign string) string {
	if region == "" || ign == "" {
		return ""
	}
	return fmt.Sprintf("https://www.leagueofgraphs.com/en/summoner/%s/%s",
		url.PathEscape(region),
		url.PathEscape(ign),
	)
}
