package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/erikfastermann/league-accounts/db"
	"github.com/erikfastermann/league-accounts/elo"
	"github.com/erikfastermann/league-accounts/handler"
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
	log.Fatal(http.ListenAndServe(":8080", h))
}
