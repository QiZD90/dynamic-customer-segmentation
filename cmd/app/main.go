package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/config"
	_ "github.com/QiZD90/dynamic-customer-segmentation/docs"
	v1 "github.com/QiZD90/dynamic-customer-segmentation/internal/controller/http/v1"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/filestorage"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/filestorage/ondisk"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository/postgres"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/timeprovider/realtimeprovider"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/userservice/usermicroservice"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// @title Dynamic Customer Segmentation
// @version 1.0
// @description Microservice for managing analytics segments

// @contact.name Elisey Puzko
// @contact.email puzko.e02@gmail.com

// @host localhost:80
// @BasePath /
func main() {
	// Parse config
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("error while parsing config")
	}

	// Create repository
	repo, err := postgres.New(cfg.Postgres.Addr, realtimeprovider.New())
	if err != nil {
		log.Fatal().Err(err).Msg("error while connecting to postgres")
	}

	// Create filestorage
	fstorage, err := ondisk.New(cfg.OnDisk.BaseURL, cfg.OnDisk.DirectoryPath, filestorage.NewTextFormatNameSupplier())
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating ondisk filestorage")
	}

	// Connect to usermicroservice
	userService, err := usermicroservice.New(cfg.UserService.BaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("error while connecting to usermicroservice")
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
	s := service.New(repo, fstorage, userService)

	// Get mux
	mux := v1.NewMux(s)

	// Start the server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info().Msgf("Listening at %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
