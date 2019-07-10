package elo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/erikfastermann/lam/db"
	"github.com/erikfastermann/lam/handler"
)

var ErrNotFound = errors.New("account not found")

func UpdateAll(ctx context.Context, db *db.DB, l *log.Logger) error {
	accs, err := db.Accounts(ctx)
	if err != nil {
		return fmt.Errorf("elo: failed reading accounts from database, %v", err)
	}
	for _, acc := range accs {
		elo, err := Get(acc.Region, acc.IGN)
		if err != nil {
			l.Print(fmt.Errorf("elo: %v", err))
			continue
		}
		if err := db.EditElo(ctx, acc.ID, elo); err != nil {
			l.Printf("elo: couldn't update elo in database (Account-ID: %d), %v", acc.ID, err)
			continue
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
		return "", fmt.Errorf("elo: failed opening URL: %s, %v", url, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("elo: failed parsing body from %s, %v", url, err)
	}
	elo := doc.Find(".leagueTier").Text()
	if elo == "" {
		return "", fmt.Errorf("elo: couldn't find .leagueTier on %s", url)
	}
	return elo, nil
}
