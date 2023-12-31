package v1

import (
	"encoding/json"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
)

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

type JsonLink struct {
	Link string `json:"link"`
}

func (j *JsonLink) Bytes() ([]byte, error) {
	return json.Marshal(j)
}

type JsonSegments struct {
	Segments []entity.Segment `json:"segments"`
}

func (j *JsonSegments) Bytes() ([]byte, error) {
	return json.Marshal(j)
}

type JsonUserSegments struct {
	Segments []entity.UserSegment `json:"segments"`
}

func (j *JsonUserSegments) Bytes() ([]byte, error) {
	return json.Marshal(j)
}

type JsonUserIDs struct {
	UserIDs []int `json:"user_ids"`
}

func (j *JsonUserIDs) Bytes() ([]byte, error) {
	return json.Marshal(j)
}
