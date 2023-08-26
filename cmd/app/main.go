package main

import (
	"fmt"
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/config"
	v1 "github.com/QiZD90/dynamic-customer-segmentation/internal/controller/http/v1"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {
	// Parse config
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Instantiate service
	s := service.New()

	// Get mux
	mux := v1.NewMux(s)

	// Start the server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info().Msgf("Listening at %s", addr)
	http.ListenAndServe(addr, mux)
}
