package entity

import "time"

type Segment struct {
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UserSegment struct {
	Slug      string     `json:"slug"`
	AddedAt   time.Time  `json:"added_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type SegmentExpiration struct {
	Slug      string     `json:"slug"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
