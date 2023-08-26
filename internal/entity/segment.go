package entity

import "time"

type Segment struct {
	Slug      string
	CreatedAt time.Time
	DeletedAt time.Time
}

type SegmentExpiration struct {
	Slug      string
	ExpiresAt time.Time
}
