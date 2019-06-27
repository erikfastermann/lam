package handler

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/erikfastermann/lam/db"
	"github.com/go-sql-driver/mysql"
)

type editPage struct {
	Title    string
	Users    []string
	Username string
	Account  db.Account
}

func accFromForm(form url.Values) (*db.Account, error) {
	acc := new(db.Account)

	formVal := func(key string) string {
		val, ok := form[key]
		if !ok {
			return ""
		}
		return val[0]
	}

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
	var ban mysql.NullTime
	if banForm == "" {
		ban = mysql.NullTime{Valid: false}
	} else {
		banTime, err := time.Parse("2006-01-02 15:04", banForm)
		if err != nil {
			return nil, err
		}
		ban = mysql.NullTime{Time: banTime, Valid: true}
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

func URLFromIGN(region, ign string) (string, error) {
	if region == "" || ign == "" {
		return "", errors.New("ign or region unset")
	}
	url, err := url.Parse(fmt.Sprintf("https://www.leagueofgraphs.com/en/summoner/%s/%s", region, ign))
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
