package v1

import "encoding/json"

type JsonResponse interface {
	Bytes() ([]byte, error)
}

type JsonError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"error_message"`
}

func (j *JsonError) Bytes() ([]byte, error) {
	return json.Marshal(j)
}

type JsonStatus struct {
	Status string `json:"status"`
}

func (j *JsonStatus) Bytes() ([]byte, error) {
	return json.Marshal(j)
}

type JsonMessage struct {
	Message string `json:"message"`
}

func (j *JsonMessage) Bytes() ([]byte, error) {
	return json.Marshal(j)
}
