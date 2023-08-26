package v1

import "github.com/QiZD90/dynamic-customer-segmentation/internal/entity"

type JsonCreateSegmentRequest struct {
	Slug string
}

type JsonDeleteSegmentRequest struct {
	Slug string
}

type JsonUserUpdateRequest struct {
	UserID         int                        `json:"user_id"`
	AddSegments    []entity.SegmentExpiration `json:"add_segments"`
	RemoveSegments []entity.SegmentExpiration `json:"remove_segments"`
}

type JsonUserSegmentsHandler struct {
	UserID int `json:"user_id"`
}
