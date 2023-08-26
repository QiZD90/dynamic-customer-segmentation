package entity

import "time"

type OperationType string

const (
	AddedOperationType   OperationType = "added"
	RemovedOperationType OperationType = "removed"
	ExpiredOperationType OperationType = "expired"
)

type Operation struct {
	UserID      int
	SegmentSlug string
	Type        OperationType
	Time        time.Time
}
