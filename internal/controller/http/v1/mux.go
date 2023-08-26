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
	mux.Get("/health", routes.HealthHandler)
	mux.MethodNotAllowed(routes.MethodNotAllowedHandler)
	mux.NotFound(routes.NotFoundHandler)

	mux.Mount("/api/v1", apiMux(routes))

	return mux
}

func apiMux(routes *Routes) http.Handler {
	mux := chi.NewMux()

	mux.Post("/segment/create", routes.SegmentCreateHandler)
	mux.Post("/segment/delete", routes.SegmentDeleteHandler)
	mux.Post("/user/update", routes.UserUpdateHandler)
	mux.Get("/user/segments", routes.UserSegmentsHandler)

	return mux
}
