package service

import (
	"context"
	"time"

	"avito/internal/domain"
	"avito/internal/repository"
)

type PullRequestService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	txMgr    repository.TransactionManager
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	txMgr repository.TransactionManager,
) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		txMgr:    txMgr,
	}
}

func (s *PullRequestService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {

	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.NewAppError(domain.ErrCodePRExists, "PR id already exists")
	}

	author, err := s.userRepo.Get(ctx, authorID)
	if err != nil {
		return nil, err
	}

	teamUsers, err := s.userRepo.GetByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}
	if len(teamUsers) == 0 {
		return nil, domain.NewAppError(domain.ErrCodeNotFound, "team not found")
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		CreatedAt:         time.Now(),
		AssignedReviewers: []string{},
	}

	err = s.prRepo.Create(ctx, tx, pr)
	if err != nil {
		return nil, err
	}

	reviewers, err := s.userRepo.GetActiveTeammates(ctx, authorID, 2)
	if err != nil {
		return nil, err
	}

	for _, r := range reviewers {
		if err := s.prRepo.AddReviewer(ctx, tx, pr.ID, r.ID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, r.ID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.Get(ctx, prID)

	if err != nil {
		return nil, err
	}

	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	pr.Status = domain.PRStatusMerged
	pr.MergedAt = &now

	if err := s.prRepo.Update(ctx, tx, pr); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()

	pr, err := s.prRepo.GetForUpdate(ctx, tx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == domain.PRStatusMerged {
		return nil, "", domain.NewAppError(domain.ErrCodePRMerged, "cannot reassign on merged PR")
	}

	isAssigned := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldUserID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", domain.NewAppError(domain.ErrCodeNotAssigned, "reviewer is not assigned to this PR")
	}

	oldReviewer, err := s.userRepo.Get(ctx, oldUserID)
	if err != nil {
		return nil, "", err
	}

	excludeIDs := []string{pr.AuthorID, oldUserID}
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID != oldUserID {
			excludeIDs = append(excludeIDs, reviewerID)
		}
	}

	newReviewer, err := s.userRepo.FindReplacementReviewer(ctx, tx, oldReviewer.TeamName, excludeIDs)
	if err != nil {
		return nil, "", err
	}

	if err := s.prRepo.RemoveReviewer(ctx, tx, prID, oldUserID); err != nil {
		return nil, "", err
	}

	if err := s.prRepo.AddReviewer(ctx, tx, prID, newReviewer.ID); err != nil {
		return nil, "", err
	}

	newReviewers := []string{}
	for _, rid := range pr.AssignedReviewers {
		if rid != oldUserID {
			newReviewers = append(newReviewers, rid)
		}
	}
	newReviewers = append(newReviewers, newReviewer.ID)
	pr.AssignedReviewers = newReviewers

	if err := tx.Commit(); err != nil {
		return nil, "", err
	}

	return pr, newReviewer.ID, nil
}
