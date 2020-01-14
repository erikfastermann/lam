package elo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/erikfastermann/lam/db"
	"github.com/erikfastermann/lam/handler"
	"golang.org/x/net/html"
)

var ErrNotFound = errors.New("account not found")

func UpdateAll(db *db.DB) error {
	ctx := context.TODO()
	accs, err := db.Accounts(ctx)
	if err != nil {
		return fmt.Errorf("failed reading accounts from database, %v", err)
	}
	for _, acc := range accs {
		time.Sleep(time.Second)
		elo, err := Get(acc.Region, acc.IGN)
		if err != nil {
			if err == ErrNotFound {
				continue
			}
			return err
		}
		if err := db.EditElo(ctx, acc.ID, elo); err != nil {
			return fmt.Errorf("couldn't update elo in database (Account-ID: %d), %v", acc.ID, err)
		}
	}
	return nil
}

func Get(region, ign string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := handler.LeagueOfGraphsURL(region, ign)
	if url == "" {
		return "", ErrNotFound
	}
	res, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed opening URL: %s, %v", url, err)
	}
	if res.StatusCode == 404 {
		return "", ErrNotFound
	}
	defer res.Body.Close()

	z := html.NewTokenizer(res.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return "", fmt.Errorf("parsing error, %v", z.Err())
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "div" || t.Data == "span" {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "leagueTier" {
						if tt := z.Next(); tt != html.TextToken {
							return "", errors.New("parsing error, structure changed")
						}
						return strings.TrimSpace(z.Token().Data), nil
					}
				}
			}
		}
	}
}
