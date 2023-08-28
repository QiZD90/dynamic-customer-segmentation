package v1

import (
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewMux(s service.Service) http.Handler {
	mux := chi.NewMux()

	routes := &Routes{s: s}
	fs := http.FileServer(http.Dir("csv"))

	mux.Use(middleware.Logger)
	mux.Get("/health", routes.HealthHandler)
	mux.Handle("/csv/*", routes.CSVOnDiskHandlerWrapper(fs))
	mux.MethodNotAllowed(routes.MethodNotAllowedHandler)
	mux.NotFound(routes.NotFoundHandler)

	mux.Get("/swagger/*", httpSwagger.Handler())

	mux.Mount("/api/v1", apiMux(routes))

	return mux
}

func apiMux(routes *Routes) http.Handler {
	mux := chi.NewMux()

	mux.Get("/segments", routes.SegmentsHandler)
	mux.Get("/segments/active", routes.SegmentsActiveHandler)
	mux.Post("/segment/create", routes.SegmentCreateHandler)
	mux.Post("/segment/create/enroll", routes.SegmentCreateEnrollHandler)
	mux.Post("/segment/delete", routes.SegmentDeleteHandler)
	mux.Post("/user/update", routes.UserUpdateHandler)
	mux.Get("/user/segments", routes.UserSegmentsHandler)
	mux.Get("/user/csv", routes.UserCSVHandler)

	return mux
}
