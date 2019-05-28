package elo

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/erikfastermann/league-accounts/db"
	"github.com/erikfastermann/league-accounts/handler"
)

func Parse(db db.Service) {
	l := log.New(os.Stderr, "elo: ", log.Ldate|log.Ltime)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for range time.NewTicker(time.Hour).C {
		accs, err := db.Accounts()
		if err != nil {
			l.Printf("failed reading accounts from database, %v", err)
			continue
		}
		for _, acc := range accs {
			url, err := handler.URLFromIGN(acc.Region, acc.IGN)
			if err != nil {
				l.Printf("couldn't create URL, %v", err)
				continue
			}
			res, err := client.Get(url)
			if err != nil {
				l.Printf("failed opening URL: %s, %v", url, err)
				continue
			}
			defer res.Body.Close()

			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				l.Printf("failed parsing body from %s, %v", url, err)
				continue
			}
			elo := doc.Find(".leagueTier").Text()
			if elo == "" {
				l.Printf("couldn't find .leagueTier on %s", url)
				continue
			}
			if err := db.EditElo(acc.ID, elo); err != nil {
				l.Printf("couldn't update elo in database (URL: %s, Account-ID: %d), %v", url, acc.ID, err)
				continue
			}
		}
	}
}
