package service

import (
	"errors"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository"
)

var (
	ErrSegmentAlreadyExists = errors.New("segment with this slug already exists")
	ErrSegmentNotFound      = errors.New("segment with this slug wasn't found")
	ErrInvalidSegmentList   = errors.New("segment list is invalid")
)

type Service struct {
	Repository repository.Repository
}

func (s *Service) CreateSegment(slug string) error {
	err := s.Repository.CreateSegment(slug)
	if errors.Is(err, repository.ErrAlreadyExists) {
		return ErrSegmentAlreadyExists
	}

	return err
}

func (s *Service) DeleteSegment(slug string) error {
	err := s.Repository.DeleteSegment(slug)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSegmentNotFound
	}

	return err
}

func (s *Service) UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error {
	if !s.ValidateSegmentLists(addSegments, removeSegments) {
		return ErrInvalidSegmentList
	}

	return s.Repository.UpdateUserSegments(userID, addSegments, removeSegments)
}

func (s *Service) GetUserActiveSegments(userID int) ([]entity.Segment, error) {
	return s.Repository.GetUserActiveSegments(userID)
}

func New(repo repository.Repository) *Service {
	return &Service{Repository: repo}
}
