package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/erikfastermann/lam/db"
	"github.com/erikfastermann/lam/elo"
	"github.com/erikfastermann/lam/handler"
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

	path := envVar("LAM_DB_PATH")
	db, err := db.Init(path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	templateGlob := envVar("LAM_TEMPLATE_GLOB")
	templates := template.Must(template.ParseGlob(templateGlob))

	go elo.Parse(db)

	h := handler.New(db, templates)
	if os.Getenv("LAM_PROD") != "" {
		domains := strings.Split(envVar("LAM_PROD_DOMAINS"), ",")
		log.Printf("Production = true, Domains: %s", domains)
		log.Fatal(http.Serve(autocert.NewListener(domains...), h))
		return
	}
	port := envVar("LAM_PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), h))
}
