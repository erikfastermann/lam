package elo

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/erikfastermann/lam/db"
	"github.com/erikfastermann/lam/handler"
)

func UpdateAll(db *db.DB, l *log.Logger) error {
	accs, err := db.Accounts()
	if err != nil {
		return fmt.Errorf("elo: failed reading accounts from database, %v", err)
	}
	for _, acc := range accs {
		elo, err := GetElo(acc.Region, acc.IGN)
		if err != nil {
			l.Print(fmt.Errorf("elo: %v", err))
			continue
		}
		if err := db.EditElo(acc.ID, elo); err != nil {
			l.Printf("elo: couldn't update elo in database (Account-ID: %d), %v", acc.ID, err)
			continue
		}
	}
	return nil
}

func GetElo(region, ign string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url, err := handler.URLFromIGN(region, ign)
	if err != nil {
		return "", fmt.Errorf("elo: couldn't create URL, %v", err)
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
