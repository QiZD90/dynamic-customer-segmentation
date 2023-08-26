package v1

import (
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/rs/zerolog/log"
)

type Routes struct {
	s *service.Service
}

// Uses manual JSON formatting since this function can be called if marshalling
// actual json error data fails
func internalServerError(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("{\"status_code\": 500, \"error_message\": \"Internal server error\"}"))
}

func respondWithJson(w http.ResponseWriter, statusCode int, j JsonResponse) {
	b, err := j.Bytes()
	if err != nil {
		log.Error().Err(err).Msg("")
		internalServerError(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func (routes *Routes) HealthHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, &JsonStatus{"OK"})
}

func (routes *Routes) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	j := &JsonError{http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed)}
	respondWithJson(w, http.StatusMethodNotAllowed, j)
}

func (routes *Routes) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	j := &JsonError{http.StatusNotFound, http.StatusText(http.StatusNotFound)}
	respondWithJson(w, http.StatusMethodNotAllowed, j)
}

func (routes *Routes) IndexHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, &JsonMessage{"CHANGE DA WORLD. MY FINAL MESSAGE. GOODBYE."})
}
