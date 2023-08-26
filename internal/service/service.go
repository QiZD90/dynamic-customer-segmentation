package service

import "github.com/QiZD90/dynamic-customer-segmentation/internal/repository"

type Service struct {
	Repository repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{Repository: repo}
}
