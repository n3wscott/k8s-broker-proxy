package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/n3wscott/k8s-broker-proxy/pkg/server"
)

type Config struct {
	Database struct {
		DSN string
	}
	Server struct {
		Addr         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}
	Bolt struct {
		File string
	}
	Auth struct {
		IntrospectURL string
	}
}

func load(file string) (cfg Config, err error) {
	cfg.Server.Addr = ":8080"
	cfg.Server.ReadTimeout = 5 * time.Second
	cfg.Server.WriteTimeout = 5 * time.Second
	cfg.Server.IdleTimeout = 5 * time.Second

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return
	}

	return
}

func main() {
	filename := flag.String("config", "config.yml", "Configuration file")
	flag.Parse()

	config, err := load(*filename)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := bolt.Open(config.Bolt.File, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := server.CreateServer(db)

	srv := &http.Server{
		ReadHeaderTimeout: config.Server.ReadTimeout,
		IdleTimeout:       config.Server.IdleTimeout,
		ReadTimeout:       config.Server.ReadTimeout,
		WriteTimeout:      config.Server.WriteTimeout,
		Addr:              config.Server.Addr,
		Handler:           s.Router,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
