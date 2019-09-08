package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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

type Config struct {
	DBPath        string
	TemplateGlob  string
	Address       string
	HTTPS         bool
	HTTPSAddress  string
	HTTPSDomain   string
	HTTPSCertKeys []CertKey
}

type CertKey struct {
	CertFile string
	KeyFile  string
}

type entry struct {
	name string
	dest *string
}

func ConfigFromEnv(prefix string) (Config, error) {
	var config Config

	err := getenv(prefix,
		entry{"DB_PATH", &config.DBPath},
		entry{"TEMPLATE_GLOB", &config.TemplateGlob},
		entry{"ADDRESS", &config.Address},
	)
	if err != nil {
		return Config{}, err
	}

	if err := getenv(prefix, entry{"HTTPS_ADDRESS", &config.HTTPSAddress}); err == nil {
		config.HTTPS = true

		var certKeys string
		err := getenv(prefix,
			entry{"HTTPS_DOMAIN", &config.HTTPSDomain},
			entry{"HTTPS_CERT_KEYS", &certKeys},
		)
		if err != nil {
			return Config{}, err
		}

		config.HTTPSCertKeys, err = parseCertKeys(certKeys)
		if err != nil {
			return Config{}, err
		}
	}

	return config, nil
}

func getenv(prefix string, entries ...entry) error {
	for _, e := range entries {
		env := prefix + "_" + e.name
		*e.dest = os.Getenv(env)
		if *e.dest == "" {
			return fmt.Errorf("env %s is empty", env)
		}
	}
	return nil
}

func parseCertKeys(certKeys string) ([]CertKey, error) {
	pairs := strings.Split(certKeys, ",")
	parsed := make([]CertKey, 0)
	for _, pair := range pairs {
		split := strings.Split(pair, ":")
		if len(split) != 2 {
			return nil, errors.New("maleformed cert-keys")
		}
		parsed = append(parsed, CertKey{
			CertFile: split[0],
			KeyFile:  split[1],
		})
	}
	return parsed, nil
}

func ListenAndServe(ctx context.Context, config Config, logger *log.Logger) error {
	h := new(handler.Handler)
	var err error
	h.DB, err = sqlite3.Init(ctx, config.DBPath)
	if err != nil {
		return err
	}
	defer h.DB.Close()

	h.Templates, err = template.ParseGlob(config.TemplateGlob)
	if err != nil {
		return err
	}

	go func() {
		duration := time.Hour
		for {
			err := elo.UpdateAll(h.DB)
			if err != nil {
				duration *= 2
				logger.Printf("elo: %v, retrying in %s", err, duration)
			}
			time.Sleep(duration)
		}
	}()

	srv := newServer(config.Address, httpwrap.LogCustom(httpwrap.HandleError(h), logger))

	if config.HTTPS {
		go func() {
			srv := newServer(config.Address,
				httpwrap.Log(http.RedirectHandler(config.HTTPSDomain, http.StatusMovedPermanently)))
			logger.Fatal(srv.ListenAndServe())
		}()

		srv.Addr = config.HTTPSAddress
		var err error
		srv.TLSConfig, err = newTLSConfig(config.HTTPSCertKeys)
		if err != nil {
			return err
		}

		listener, err := tls.Listen("tcp", config.HTTPSAddress, srv.TLSConfig)
		if err != nil {
			return err
		}

		logger.Printf("server: listening on address %s (https)", config.HTTPSAddress)
		logger.Printf("server: redirecting http (address: %s) to https", config.Address)
		return srv.Serve(listener)
	}

	logger.Printf("server: listening on address %s (http)", config.Address)
	return srv.ListenAndServe()
}

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        h,
		MaxHeaderBytes: 1 << 20,
	}
}

func newTLSConfig(certKeys []CertKey) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}
	for _, ck := range certKeys {
		keyPair, err := tls.LoadX509KeyPair(ck.CertFile, ck.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, keyPair)
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}
