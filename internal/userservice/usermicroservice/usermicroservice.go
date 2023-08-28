package usermicroservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type UserMicroservice struct {
	BaseURL string
}

type JsonUsersRandomRequest struct {
	Percent int `json:"percent"`
}

type JsonUsersRandomResponse struct {
	UserIDs []int `json:"user_ids"`
}

func (u *UserMicroservice) GetRandomUsers(percent int) ([]int, error) {
	apiUrl, err := url.JoinPath(u.BaseURL, "/api/v1/users/random")
	if err != nil {
		return nil, fmt.Errorf("usermicroservice.GetRandomUsers() - url.JoinPath(): %w", err)
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(JsonUsersRandomRequest{percent}); err != nil {
		return nil, fmt.Errorf("usermicroservice.GetRandomUsers() - json.Encoder.Encode(): %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, apiUrl, &b)
	if err != nil {
		return nil, fmt.Errorf("usermicroservice.GetRandomUsers() - http.NewRequest(): %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("usermicroservice.GetRandomUsers() - http.DefaultClient.Do(): %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usermicroservice.GetRandomUsers - http.Get(): status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var j JsonUsersRandomResponse
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return nil, fmt.Errorf("usermicroservice.GetRandomUsers - unmarshall response error: %w", err)
	}

	return j.UserIDs, nil
}

func New(baseURL string) (*UserMicroservice, error) {
	userService := &UserMicroservice{BaseURL: baseURL}

	// ping it for good measure :P
	pingUrl, err := url.JoinPath(userService.BaseURL, "/health")
	if err != nil {
		return nil, fmt.Errorf("usermicroservice.New() - url.JoinPath(): %w", err)
	}

	resp, err := http.Get(pingUrl)
	if err != nil {
		return nil, fmt.Errorf("usermicroservice.New() - http.Get(): %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usermicroservice.New(): ping failed")
	}

	return userService, nil
}
