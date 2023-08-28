// Mock implementation of microservice that handles user database
// It's quick, dirty and hacky so don't look at it too much

package main

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

type JsonUsersRandomRequest struct {
	Percent int `json:"percent"`
}

type JsonUsersRandomResponse struct {
	UserIDs []int `json:"user_ids"`
}

func UsersRandomHandler(w http.ResponseWriter, r *http.Request) {
	var j JsonUsersRandomRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Error().Err(err).Msg("")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"status_code\": 400, \"error_message\": \"Error while unmarshalling request JSON\"}"))
		return
	}

	takenUsers := make(map[int]struct{})
	usersCnt := j.Percent * 3
	for i := 0; i < usersCnt; i++ {
		randomID := int(1000 + rand.Uint32()%300)
		_, ok := takenUsers[randomID]

		for ok {
			randomID = int(1000 + rand.Uint32()%300)
			_, ok = takenUsers[randomID]
		}

		takenUsers[randomID] = struct{}{}
	}

	users := make([]int, 0, len(takenUsers))
	for k := range takenUsers {
		users = append(users, k)
	}

	response := JsonUsersRandomResponse{users}
	b, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"status_code\": 500, \"error_message\": \"Internal server error\"}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func main() {
	mux := chi.NewMux()

	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"status\": \"OK\"}"))
	})
	mux.Get("/api/v1/users/random", UsersRandomHandler)

	log.Info().Msg("Listening at :80")
	if err := http.ListenAndServe(":80", mux); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
