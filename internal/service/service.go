package service

import (
	"errors"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository"
)

var (
	ErrSegmentAlreadyExists  = errors.New("segment with this slug already exists")
	ErrSegmentNotFound       = errors.New("segment with this slug wasn't found")
	ErrSegmentAlreadyDeleted = errors.New("segment with this slug is already deleted")
	ErrInvalidSegmentList    = errors.New("segment list is invalid")
)

type Service interface {
	// CreateSegment creates a segment with specified slug.
	// If there is a segment (active or deleted) with this slug already, returns `ErrSegmentAlreadyExists`
	CreateSegment(slug string) error

	CreateSegmentAndEnrollPercent(slug string, percent int) error

	// DeleteSegment marks segment as deleted and marks all records with it as removed
	// Returns `ErrSegmentNotFound` if there is no segment by this slug
	DeleteSegment(slug string) error

	// GetAllActiveSegments returns all active segments
	GetAllActiveSegments() ([]entity.Segment, error)

	// GetAllActiveSegments returns all segments, active or not
	GetAllSegments() ([]entity.Segment, error)

	// UpdateUserSegments adds and removes segments to/from user with expiration date
	// If user is already in the segment that you want to add, ignores it.
	// If user doesn't have the segment that you want to remove, ignores it.
	// If segment any of the segments don't exist or was deleted returns `ErrSegmentNotFound` and `ErrSegmentAlreadyDeleted`
	UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error

	// GetActiveUserSegments returns active (not removed and not expired) segments that user is in
	GetActiveUserSegments(userID int) ([]entity.UserSegment, error)

	// DumpHistory returns all operations related to given users that occurred in specified time span
	// Returns a download link for a CSV file with this data
	DumpHistoryCSV(userIDs []int, timeFrom time.Time, timeTo time.Time) (string, error)
}

type SegmentationService struct {
	Repository repository.Repository
}

func (s *SegmentationService) CreateSegment(slug string) error {
	err := s.Repository.CreateSegment(slug)
	if errors.Is(err, repository.ErrSegmentAlreadyExists) {
		return ErrSegmentAlreadyExists
	}

	return err
}

func (s *SegmentationService) DeleteSegment(slug string) error {
	err := s.Repository.DeleteSegment(slug)
	if errors.Is(err, repository.ErrSegmentNotFound) {
		return ErrSegmentNotFound
	} else if errors.Is(err, repository.ErrSegmentAlreadyDeleted) {
		return ErrSegmentAlreadyDeleted
	}

	return err
}

func (s *SegmentationService) CreateSegmentAndEnrollPercent(slug string, percent int) error { // TODO:
	return nil
}

func (s *SegmentationService) GetAllActiveSegments() ([]entity.Segment, error) {
	return s.Repository.GetAllActiveSegments()
}

func (s *SegmentationService) GetAllSegments() ([]entity.Segment, error) {
	return s.Repository.GetAllSegments()
}

func (s *SegmentationService) UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error {
	if !ValidateSegmentLists(addSegments, removeSegments) {
		return ErrInvalidSegmentList
	}

	err := s.Repository.UpdateUserSegments(userID, addSegments, removeSegments)
	if errors.Is(err, repository.ErrSegmentNotFound) {
		return ErrSegmentNotFound
	} else if errors.Is(err, repository.ErrSegmentAlreadyDeleted) {
		return ErrSegmentAlreadyDeleted
	}

	return err
}

func (s *SegmentationService) GetActiveUserSegments(userID int) ([]entity.UserSegment, error) {
	return s.Repository.GetActiveUserSegments(userID)
}

func (s *SegmentationService) DumpHistoryCSV(userIDs []int, timeFrom time.Time, timeTo time.Time) (string, error) { // TODO:
	return "", nil
}

func New(repo repository.Repository) *SegmentationService {
	return &SegmentationService{Repository: repo}
}
