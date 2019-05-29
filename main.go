package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/erikfastermann/league-accounts/db"
	"github.com/erikfastermann/league-accounts/elo"
	"github.com/erikfastermann/league-accounts/handler"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	envVar := func(env string) string {
		str := os.Getenv(env)
		if str == "" {
			log.Fatalln("Environment variable", env, "is empty or doesn't exist")
		}
		return str
	}

	user := envVar("MYSQL_USER")
	password := envVar("MYSQL_PASSWORD")
	address := envVar("LEAGUE_ACCS_DB_ADDRESS")
	name := envVar("MYSQL_DATABASE")
	db, err := db.New(user, password, address, name)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	templateDir := envVar("LEAGUE_ACCS_TEMPLATE_DIR")
	templates := template.Must(template.ParseGlob(templateDir))

	go elo.Parse(db)

	h := handler.New(db, templates)
	if os.Getenv("LEAGUE_ACCS_PROD") != "" {
		domains := strings.Split(envVar("LEAGUE_ACCS_PROD_DOMAINS"), ",")
		log.Printf("Production = true, Domains: %s", domains)
		log.Fatal(http.Serve(autocert.NewListener(domains...), h))
		return
	}
	port := envVar("LEAGUE_ACCS_PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), h))
}
