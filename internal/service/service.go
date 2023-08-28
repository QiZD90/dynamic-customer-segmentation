package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/filestorage"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/userservice"
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

	// CreateSegmentAndEnrollPercent creates segment using CreateSegment, gets random users
	// through UserService and then tries to add the segment to them.
	// Returns ids of selected users (they may or may not have got the segment added)
	// May return `ErrSegmentNotFound` or `ErrSegmentAlreadyExists`
	CreateSegmentAndEnrollPercent(slug string, percent int) ([]int, error)

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
	DumpHistoryCSV(userID int, timeFrom time.Time, timeTo time.Time) (string, error)
}

type SegmentationService struct {
	Repository  repository.Repository
	FileStorage filestorage.FileStorage
	UserService userservice.UserService
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

func (s *SegmentationService) CreateSegmentAndEnrollPercent(slug string, percent int) ([]int, error) {
	if err := s.CreateSegment(slug); err != nil {
		if errors.Is(err, repository.ErrSegmentAlreadyExists) {
			return nil, ErrSegmentAlreadyExists
		}

		return nil, err
	}

	userIDs, err := s.UserService.GetRandomUsers(percent)
	if err != nil {
		return nil, err
	}

	if err := s.Repository.AddSegmentToUsers(slug, userIDs); err != nil {
		if errors.Is(err, repository.ErrSegmentAlreadyExists) {
			return nil, ErrSegmentAlreadyExists
		} else if errors.Is(err, repository.ErrSegmentNotFound) {
			return nil, ErrSegmentNotFound
		}

		return nil, err
	}

	return userIDs, nil
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

func (s *SegmentationService) DumpHistoryCSV(userID int, timeFrom time.Time, timeTo time.Time) (string, error) {
	operations, err := s.Repository.DumpHistory(userID, timeFrom, timeTo)
	if err != nil {
		return "", err
	}

	csv := s.generateCSVString(userID, operations)
	csvURL, err := s.FileStorage.StoreCSV(csv, userID, timeFrom, timeTo)
	return csvURL, err
}

func (s *SegmentationService) generateCSVString(userID int, operations []entity.Operation) string {
	sb := strings.Builder{}

	for _, o := range operations {
		sb.WriteString(fmt.Sprintf("%d;%s;%s;%s\n", userID, o.SegmentSlug, o.Type, o.Time))
	}

	return sb.String()
}

func New(repo repository.Repository, fstorage filestorage.FileStorage, userService userservice.UserService) *SegmentationService {
	return &SegmentationService{Repository: repo, FileStorage: fstorage, UserService: userService}
}
