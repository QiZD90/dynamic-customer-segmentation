package v1

import (
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewMux(s *service.Service) http.Handler {
	mux := chi.NewMux()

	routes := &Routes{s: s}
	mux.Use(middleware.Logger)
	mux.Get("/", routes.IndexHandler)
	mux.MethodNotAllowed(routes.MethodNotAllowedHandler)
	mux.NotFound(routes.NotFoundHandler)

	return mux
}
