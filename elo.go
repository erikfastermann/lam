package main

import (
    "net/http"
    "time"
    "log"
    "net/url"
    "fmt"
    "github.com/PuerkitoBio/goquery"
)

func elo() {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    for range time.NewTicker(time.Hour).C {
        accs, err := allAccounts()
        if err != nil {
            log.Println("WEB-PARSER: ERROR reading from database.", err)
            return
        }
        for _, acc := range accs {
            url, err := url.Parse(fmt.Sprintf("https://www.leagueofgraphs.com/en/summoner/%s/%s", acc.Region, acc.Ign))
            if err != nil {
                log.Println("WEB-PARSER: ERROR escaping", url, err)
                continue
            }
            link := url.String()

            res, err := client.Get(link)
            if err != nil {
                log.Println("WEB-PARSER: ERROR opening", link, err)
                continue
            }
            defer res.Body.Close()

            doc, err := goquery.NewDocumentFromReader(res.Body)
            if err != nil {
                log.Println("WEB-PARSER: ERROR parsing", link, err)
                continue
            }
            leagueTier := doc.Find(".leagueTier").Text()
            if leagueTier == "" {
                log.Println("WEB-PARSER: ERROR finding .leagueTier", link)
                continue
            }

            tokenPrep, err := db.Prepare("UPDATE accounts SET Elo=? WHERE ID=?")
            if err != nil {
                log.Println("WEB-PARSER: FAILED preparing Elo", leagueTier, "for Account", acc.Ign, err)
                continue
            }
            _, err = tokenPrep.Exec(leagueTier, acc.ID)
            if err != nil {
                log.Println("WEB-PARSER: FAILED storing Elo", leagueTier, "for Account", acc.Ign, err)
                continue
            }
            log.Println("WEB-PARSER: SUCCESS storing Elo:", leagueTier, "for Account", acc.Ign)
        }
    }
}

