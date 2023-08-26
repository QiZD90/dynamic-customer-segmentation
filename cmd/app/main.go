package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/config"
	v1 "github.com/QiZD90/dynamic-customer-segmentation/internal/controller/http/v1"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository/postgres"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

func main() {
	// Parse config
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Create repository
	repo, err := postgres.New(cfg.Postgres.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Migrate up to date
	log.Info().Msg("Starting migrations...")
	m, err := migrate.New("file://migrations", cfg.Postgres.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("Already up to date")
		} else {
			log.Fatal().Err(err).Msg("")
		}
	}

	if src_err, db_err := m.Close(); src_err != nil || db_err != nil {
		log.Fatal().AnErr("src_err", src_err).AnErr("db_err", db_err).Msg("")
	}

	// Instantiate service
	s := service.New(repo)

	// Get mux
	mux := v1.NewMux(s)

	// Start the server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info().Msgf("Listening at %s", addr)
	http.ListenAndServe(addr, mux)
}
