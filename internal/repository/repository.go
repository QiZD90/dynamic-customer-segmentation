package repository

import (
	"errors"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
)

// !!NOTE!!: these errors are only used by service package that substitutes
// them for its own equivalents
var (
	ErrSegmentAlreadyExists  = errors.New("segment with this slug already exists")
	ErrSegmentAlreadyDeleted = errors.New("segment with this slug is already deleted")
	ErrSegmentNotFound       = errors.New("segment with this slug doesn't exist")
)

type Repository interface {
	CreateSegment(slug string) error
	AddSegmentToUsers(slug string, userIDs []int) error
	DeleteSegment(slug string) error
	GetAllActiveSegments() ([]entity.Segment, error)
	GetAllSegments() ([]entity.Segment, error)

	// !!NOTE!!: behaviour in case of duplicate entries in slices or an entry
	// being in both slices is intentionally undefined
	UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error

	GetActiveUserSegments(userID int) ([]entity.UserSegment, error)

	// DumpHistory returns all operations related to a given user that occurred in specified time span
	// sorted by operation time
	DumpHistory(userID int, timeFrom time.Time, timeTo time.Time) ([]entity.Operation, error)
}
