package v1

import "github.com/QiZD90/dynamic-customer-segmentation/internal/entity"

type JsonCreateSegmentRequest struct {
	Slug string
}

type JsonSegmentCreateAndEnroll struct {
	Slug    string
	Percent int
}

type JsonDeleteSegmentRequest struct {
	Slug string
}

type JsonUserUpdateRequest struct {
	UserID         int                        `json:"user_id"`
	AddSegments    []entity.SegmentExpiration `json:"add_segments"`
	RemoveSegments []entity.SegmentExpiration `json:"remove_segments"`
}

type JsonUserSegmentsHandlerRequest struct {
	UserID int `json:"user_id"`
}

type JsonDate struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

type JsonUserCSVRequest struct {
	UserID   int      `json:"user_id"`
	FromDate JsonDate `json:"from"`
	ToDate   JsonDate `json:"to"`
}
