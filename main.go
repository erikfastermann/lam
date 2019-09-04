package main

import (
	"context"
	"crypto/tls"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/erikfastermann/httpwrap"
	"github.com/erikfastermann/lam/db/sqlite3"
	"github.com/erikfastermann/lam/elo"
	"github.com/erikfastermann/lam/handler"
)

func main() {
	path := getenv("LAM_DB_PATH")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sqlite3.Init(ctx, path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	templateGlob := getenv("LAM_TEMPLATE_GLOB")
	templates := template.Must(template.ParseGlob(templateGlob))

	l := log.New(os.Stderr, "", log.Ldate|log.Ltime)

	go func() {
		duration := time.Hour
		for {
			err := elo.UpdateAll(db)
			if err != nil {
				duration *= 2
				l.Printf("elo: %v, retrying in %s", err, duration)
			}
			time.Sleep(duration)
		}
	}()

	h := httpwrap.LogCustom(httpwrap.HandleError(&handler.Handler{
		DB:        db,
		Templates: templates,
	}), l)

	addr := getenv("LAM_ADDRESS")
	if tlsAddr := os.Getenv("LAM_HTTPS_ADDRESS"); tlsAddr != "" {
		url := getenv("LAM_HTTPS_DOMAIN")
		go func() {
			srv := newServer(addr, redirectToHTTPS(url, l))
			log.Fatal(srv.ListenAndServe())
		}()

		srv := newServer(tlsAddr, h)
		var err error
		srv.TLSConfig, err = newTLSConfig(getenv("LAM_HTTPS_CERTS"), getenv("LAM_HTTPS_KEYS"))
		if err != nil {
			log.Fatal(err)
		}
		listener, err := tls.Listen("tcp", tlsAddr, srv.TLSConfig)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("server: listening on address %s (https)", tlsAddr)
		log.Printf("server: redirecting http (address: %s) to https", addr)
		log.Fatal(srv.Serve(listener))
	}

	srv := newServer(addr, h)
	log.Printf("server: listening on address %s (http)", addr)
	log.Fatal(srv.ListenAndServe())
}

func getenv(env string) string {
	str := os.Getenv(env)
	if str == "" {
		log.Fatalln("Environment variable", env, "is empty or doesn't exist")
	}
	return str
}

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        h,
		MaxHeaderBytes: 1 << 20,
	}
}

func newTLSConfig(certFiles, keyFiles string) (*tls.Config, error) {
	certs := strings.Split(certFiles, ",")
	keys := strings.Split(keyFiles, ",")
	if len(certs) != len(keys) {
		return nil, errors.New("different number of key and cert files")
	}
	if certs[0] == "" || keys[0] == "" {
		return nil, errors.New("no key pairs supplied")
	}
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}
	for i, cert := range certs {
		keyPair, err := tls.LoadX509KeyPair(cert, keys[i])
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, keyPair)
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func redirectToHTTPS(url string, l *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		l.Printf("redirect %s to %s", r.URL.Path, url)
	}
}
