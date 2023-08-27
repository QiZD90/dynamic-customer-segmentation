package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/rs/zerolog/log"
)

type Routes struct {
	s service.Service
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

func (routes *Routes) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	j := &JsonError{http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed)}
	respondWithJson(w, http.StatusMethodNotAllowed, j)
}

func (routes *Routes) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	j := &JsonError{http.StatusNotFound, http.StatusText(http.StatusNotFound)}
	respondWithJson(w, http.StatusMethodNotAllowed, j)
}

// GET /health
func (routes *Routes) HealthHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, &JsonStatus{"OK"})
}

// GET /segments/active
func (routes *Routes) SegmentsActiveHandler(w http.ResponseWriter, r *http.Request) {
	segments, err := routes.s.GetAllActiveSegments()
	if err != nil {
		log.Error().Err(err).Msg("")
		internalServerError(w)

		return
	}

	respondWithJson(w, http.StatusOK, &JsonSegments{segments})
}

// GET /segments/active
func (routes *Routes) SegmentsHandler(w http.ResponseWriter, r *http.Request) {
	segments, err := routes.s.GetAllSegments()
	if err != nil {
		log.Error().Err(err).Msg("")
		internalServerError(w)

		return
	}

	respondWithJson(w, http.StatusOK, &JsonSegments{segments})
}

// POST /segment/create
func (routes *Routes) SegmentCreateHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonCreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}
	defer r.Body.Close()

	if err := routes.s.CreateSegment(j.Slug); err != nil {
		log.Error().Err(err).Msg("")

		if errors.Is(err, service.ErrSegmentAlreadyExists) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment already exists"})
		} else {
			internalServerError(w)
		}

		return
	}

	respondWithJson(w, http.StatusOK, &JsonStatus{"OK"})
}

// POST /segment/delete
func (routes *Routes) SegmentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonDeleteSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}
	defer r.Body.Close()

	if err := routes.s.DeleteSegment(j.Slug); err != nil {
		log.Error().Err(err).Msg("")

		if errors.Is(err, service.ErrSegmentNotFound) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment wasn't found"})
		} else if errors.Is(err, service.ErrSegmentAlreadyDeleted) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment is already deleted"})
		} else {
			internalServerError(w)
		}

		return
	}

	respondWithJson(w, http.StatusOK, &JsonStatus{"OK"})
}

// POST /user/update
func (routes *Routes) UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonUserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})
		return
	}
	defer r.Body.Close()

	if err := routes.s.UpdateUserSegments(j.UserID, j.AddSegments, j.RemoveSegments); err != nil {
		log.Error().Err(err).Msg("")

		if errors.Is(err, service.ErrInvalidSegmentList) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Supplied segment lists are invalid"})
		} else if errors.Is(err, service.ErrSegmentNotFound) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment wasn't found"})
		} else if errors.Is(err, service.ErrSegmentAlreadyDeleted) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment is already deleted"})
		} else {
			internalServerError(w)
		}

		return
	}

	respondWithJson(w, http.StatusOK, &JsonStatus{"OK"})
}

// GET /user/segments
func (routes *Routes) UserSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonUserSegmentsHandler
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}
	defer r.Body.Close()

	segments, err := routes.s.GetActiveUserSegments(j.UserID)
	if err != nil {
		log.Error().Err(err).Msg("")
		internalServerError(w)

		return
	}

	respondWithJson(w, http.StatusOK, &JsonUserSegments{segments})
}
