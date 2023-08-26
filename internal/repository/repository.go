package repository

import (
	"errors"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
)

var (
	ErrSegmentAlreadyExists = errors.New("segment with this slug already exists")
	ErrNoSuchSegment        = errors.New("segment with this slug doesn't exist")
)

type Repository interface {
	// CreateSegment creates a segment with specified slug.
	// If there is a segment with this slug already, returns `ErrSegmentAlreadyExists`
	CreateSegment(slug string) error

	// CreateSegmentAndEnroll creates a segment with specified slug and adds it to users with specified IDs
	CreateSegmentAndEnroll(slug string, userIDs []int) error

	// DeleteSegment marks segment as deleted and marks all records with it as deleted
	// Returns `ErrNoSuchSegment` if there is no segment by this slug
	DeleteSegment(slug string) error

	// UpdateUserSegments adds and removes segments to/from user with expiration date
	// !!NOTE!!: behaviour in case of duplicate entries in slices or an entry
	// being in both slices is intentionally undefined
	UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error

	// GetUserActiveSegments returns active (not deleted and not expired) segments that user is in
	GetUserActiveSegments(userID int) ([]entity.Segment, error)

	// DumpHistory returns all operations related to given users that occurred in specified time span
	DumpHistory(userIDs []int, timeFrom time.Time, timeTo time.Time) ([]entity.Operation, error)
}
