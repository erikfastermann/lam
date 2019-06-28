package main

import (
	"crypto/tls"
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/erikfastermann/lam/db"
	"github.com/erikfastermann/lam/elo"
	"github.com/erikfastermann/lam/handler"
)

func main() {
	path := getenv("LAM_DB_PATH")
	db, err := db.Init(path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	templateGlob := getenv("LAM_TEMPLATE_GLOB")
	templates := template.Must(template.ParseGlob(templateGlob))

	go elo.Parse(db)

	h := handler.New(db, templates)
	port := getenv("LAM_PORT")
	if tlsPort := os.Getenv("LAM_HTTPS_PORT"); tlsPort != "" {
		domain := getenv("LAM_HTTPS_DOMAIN")
		go func() {
			srv := newServer(port, redirectToHTTPS(domain, tlsPort))
			log.Fatal(srv.ListenAndServe())
		}()
		srv := newServer(tlsPort, h)
		var err error
		srv.TLSConfig, err = newTLSConfig(getenv("LAM_HTTPS_CERTS"), getenv("LAM_HTTPS_KEYS"))
		if err != nil {
			log.Fatal(err)
		}
		listener, err := tls.Listen("tcp", ":"+tlsPort, srv.TLSConfig)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("server: listening on port %s (https)", tlsPort)
		log.Printf("server: redirecting http (port: %s) to https", port)
		log.Fatal(srv.Serve(listener))
		return
	}

	srv := newServer(port, h)
	log.Printf("server: listening on port %s (http)", port)
	log.Fatal(srv.ListenAndServe())
}

func getenv(env string) string {
	str := os.Getenv(env)
	if str == "" {
		log.Fatalln("Environment variable", env, "is empty or doesn't exist")
	}
	return str
}

func newServer(port string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:           ":" + port,
		Handler:        h,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
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

func redirectToHTTPS(host, tlsPort string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		u.Host = net.JoinHostPort(host, tlsPort)
		u.Scheme = "https"
		urlStr := u.String()
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
		log.Printf("redirect: %s to %s (https)", r.URL.Path, urlStr)
	}
}
