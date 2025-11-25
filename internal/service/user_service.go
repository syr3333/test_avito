package service

import (
	"context"

	"avito/internal/domain"
	"avito/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.SetActive(ctx, userID, isActive); err != nil {
		return nil, err
	}

	user, err = s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetReviewPRs(ctx context.Context, userID string, prRepo repository.PullRequestRepository) ([]domain.PullRequestShort, error) {
	// Check if user exists
	_, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prRepo.GetByReviewer(ctx, userID)
}
