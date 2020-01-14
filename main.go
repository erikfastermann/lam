package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/erikfastermann/httpwrap"
	"github.com/erikfastermann/lam/db"
	"github.com/erikfastermann/lam/elo"
	"github.com/erikfastermann/lam/handler"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	type entry struct {
		name string
		dest *string
	}

	var addr, https, domain, cert, key string
	var users, accs, ctr string
	var tmplt string
	for _, e := range []entry{
		{"ADDRESS", &addr},
		{"HTTPS_ADDRESS", &https},
		{"DOMAIN", &domain},
		{"CERT", &cert},
		{"KEY", &key},

		{"USERS", &users},
		{"ACCOUNTS", &accs},
		{"CTR", &ctr},

		{"TEMPLATE_GLOB", &tmplt},
	} {
		env := "LAM_" + e.name
		*e.dest = os.Getenv(env)
		if *e.dest == "" {
			return fmt.Errorf("env %s is empty", env)
		}
	}

	split := strings.Split(users, ":")
	if len(split)%2 != 0 {
		return fmt.Errorf("not every user has a password set")
	}
	u := make([]*handler.User, 0)
	for i := 0; i < len(split); i += 2 {
		u = append(u, &handler.User{
			Username: split[i],
			Password: split[i+1],
		})
	}

	h := &handler.Handler{
		Users: u,
	}

	var err error
	h.DB, err = db.Init(accs, ctr)
	if err != nil {
		return err
	}
	defer h.DB.Close()

	h.Templates, err = template.ParseGlob(tmplt)
	if err != nil {
		return err
	}

	go func() {
		duration := 24 * time.Hour
		l := log.New(os.Stderr, "ERROR ", log.LstdFlags)
		for {
			if err := elo.UpdateAll(h.DB); err != nil {
				l.Printf("elo: %v, retrying in %s", err, duration)
			}
			time.Sleep(duration)
		}
	}()

	go func() {
		srv := newServer(addr, httpwrap.Log(http.RedirectHandler(domain, http.StatusMovedPermanently)))
		log.Fatal(srv.ListenAndServe())
	}()

	srv := newServer(https, httpwrap.Log(httpwrap.HandleError(h)))
	log.Printf("server: listening on address %s (https)", https)
	log.Printf("server: redirecting http (address: %s) to %s", addr, domain)
	return srv.ListenAndServeTLS(cert, key)
}

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        h,
		MaxHeaderBytes: 1 << 20,
	}
}
