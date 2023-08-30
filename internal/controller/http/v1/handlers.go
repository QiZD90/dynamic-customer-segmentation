package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
// @Summary Health check
// @Produce json
// @Success 200 {object} v1.JsonStatus
// @Router /health [get]
func (routes *Routes) HealthHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, &JsonStatus{"OK"})
}

// GET /csv/*
// @Summary Get CSV file
// @Description Get static CSV file stored on disk
// @Router /csv/{fname} [get]
func (routes *Routes) CSVOnDiskHandlerWrapper(fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/csv")
		http.StripPrefix("/csv/", fs).ServeHTTP(w, r)
	})
}

// GET /segments/active
// @Summary Get all active segments
// @Description Get all active (not deleted) segments
// @Produce json
// @Success 200 {object} v1.JsonSegments
// @Router /api/v1/segments/active [get]
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
// @Summary Get all segments
// @Description Get all segments (even deleted)
// @Produce json
// @Success 200 {object} v1.JsonSegments
// @Router /api/v1/segments [get]
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
// @Summary Create new segment
// @Description Create new segment with given slug. If there is already active segment with this slug,
// @Description or if there was a segment with this slug but it has been deleted, responds with an error and 400 status code
// @Accept json
// @Produce json
// @Param input body v1.JsonCreateSegmentRequest true "input"
// @Success 200 {object} v1.JsonStatus
// @Failure 400 {object} v1.JsonError
// @Failure 500 {object} v1.JsonError
// @Router /api/v1/segment/create [post]
func (routes *Routes) SegmentCreateHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonCreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}

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

// POST /segment/create/enroll
// @Summary Creates new segment and adds it to randomly selected users
// @Description Creates new segment with given slug. If there is already active segment with this slug,
// @Description or if there was a segment with this slug but it has been deleted, responds with an error and 400 status code
// @Description Get a percent of randomly selected users from user DB service and tries to add the newly created segment to them.
// @Accept json
// @Produce json
// @Param input body v1.JsonSegmentCreateAndEnroll true "input"
// @Success 200 {object} v1.JsonUserIDs "IDs of users that were selected"
// @Failure 400 {object} v1.JsonError
// @Failure 500 {object} v1.JsonError
// @Router /api/v1/segment/create/enroll [post]
func (routes *Routes) SegmentCreateEnrollHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonSegmentCreateAndEnroll
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}

	if j.Percent < 0 || j.Percent > 100 {
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Invalid percent value"})

		return
	}

	userIDs, err := routes.s.CreateSegmentAndEnrollPercent(j.Slug, j.Percent)
	if err != nil {
		log.Error().Err(err).Msg("")

		if errors.Is(err, service.ErrSegmentAlreadyExists) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment already exists"})
		} else if errors.Is(err, service.ErrSegmentNotFound) {
			respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Segment wasn't found"})
		} else {
			internalServerError(w)
		}

		return
	}

	respondWithJson(w, http.StatusOK, &JsonUserIDs{UserIDs: userIDs})
}

// POST /segment/delete
// @Summary Delete a segment
// @Description Marks a segment by this slug as deleted. If there is no segment like this, or if was already deleted,
// @Description responds with an error and 400 status code
// @Accept json
// @Produce json
// @Param input body v1.JsonDeleteSegmentRequest true "input"
// @Success 200 {object} v1.JsonStatus
// @Failure 400 {object} v1.JsonError
// @Failure 500 {object} v1.JsonError
// @Router /api/v1/segment/delete [post]
func (routes *Routes) SegmentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonDeleteSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}

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
// @Summary Add and remove segments from user
// @Description Tries to add and remove segments from user. If any of the specified segments are not active
// @Description or if any of the lists contains same segment twice or if both list contain the same segment
// @Description responds with an error and 400 status code.
// @Description You can specify expiry date for segments. This field is ignored in segments in remove list.
// @Description If you try add a segment to a user that already has it or you try to remove it from a user
// @Description that doesn't have it then that segment is skipped. Note, that if you try to modify expiry
// @Description date of an active segment, the correct way to do it is to remove it and then add a new one.
// @Accept json
// @Produce json
// @Param input body v1.JsonUserUpdateRequest true "input"
// @Success 200 {object} v1.JsonStatus
// @Failure 400 {object} v1.JsonError
// @Failure 500 {object} v1.JsonError
// @Router /api/v1/user/update [post]
func (routes *Routes) UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonUserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})
		return
	}

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
// @Summary Get user's active segments
// @Accept json
// @Produce json
// @Param input body v1.JsonUserSegmentsHandlerRequest true "input"
// @Success 200 {object} v1.JsonUserSegments
// @Failure 400 {object} v1.JsonError
// @Failure 500 {object} v1.JsonError
// @Router /api/v1/user/segments [get]
func (routes *Routes) UserSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonUserSegmentsHandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}

	segments, err := routes.s.GetActiveUserSegments(j.UserID)
	if err != nil {
		log.Error().Err(err).Msg("")
		internalServerError(w)

		return
	}

	respondWithJson(w, http.StatusOK, &JsonUserSegments{segments})
}

// GET /user/csv
// @Summary Generate CSV report on user's segment history
// @Description Generate CSV report file on user's segment history and uploads it to service's configured file storage service.
// @Description Note thah `month` param in date is an integer that ranges from 1 (january) to 12 (december)
// @Description Also note that the specified range includes the "from" date but excludes the "to" date
// @Accept json
// @Produce json
// @Param input body v1.JsonUserCSVRequest true "input"
// @Success 200 {object} v1.JsonLink
// @Failure 400 {object} v1.JsonError
// @Failure 500 {object} v1.JsonError
// @Router /api/v1/user/csv [get]
func (routes *Routes) UserCSVHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonUserCSVRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Error while unmarshalling request JSON"})

		return
	}

	if j.FromDate.Month < 1 || j.FromDate.Month > 12 {
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Incorrect month in from date"})
		return
	}

	if j.ToDate.Month < 1 || j.ToDate.Month > 12 {
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "Incorrect month in to date"})
		return
	}

	if j.FromDate.Month+j.FromDate.Year*12 > j.ToDate.Month+j.ToDate.Year*12 {
		respondWithJson(w, http.StatusBadRequest, &JsonError{http.StatusBadRequest, "From date is later than to date"})
		return
	}

	fromTime := time.Date(j.FromDate.Year, time.Month(j.FromDate.Month), 1, 0, 0, 0, 0, time.UTC)
	toTime := time.Date(j.ToDate.Year, time.Month(j.ToDate.Month), 1, 0, 0, 0, 0, time.UTC)

	link, err := routes.s.DumpHistoryCSV(j.UserID, fromTime, toTime)
	if err != nil {
		log.Error().Err(err).Msg("")
		internalServerError(w)
		return
	}

	respondWithJson(w, http.StatusOK, &JsonLink{link})
}
