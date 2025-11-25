package service

import (
	"context"
	"math/rand"
	"time"

	"avito/internal/domain"
	"avito/internal/repository"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
	txMgr    repository.TransactionManager
}

func NewTeamService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	txMgr repository.TransactionManager,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		prRepo:   prRepo,
		txMgr:    txMgr,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) error {
	exists, err := s.teamRepo.Exists(ctx, team.Name)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewAppError(domain.ErrCodeTeamExists, "team_name already exists")
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.teamRepo.Create(ctx, tx, team.Name); err != nil {
		return err
	}

	for _, member := range team.Members {
		user := &domain.User{
			ID:       member.UserID,
			Username: member.Username,
			TeamName: team.Name,
			IsActive: member.IsActive,
		}
		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	return s.teamRepo.Get(ctx, teamName)
}

func (s *TeamService) MassDeactivateUsers(ctx context.Context, teamName string, userIDs []string) error {
	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.userRepo.DeactivateMany(ctx, tx, userIDs); err != nil {
		return err
	}

	assignments, err := s.prRepo.GetOpenAssignmentsByReviewers(ctx, tx, userIDs)
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		return tx.Commit()
	}

	activeUsers, err := s.userRepo.GetActiveUsersByTeam(ctx, tx, teamName)
	if err != nil {
		return err
	}

	prIDsSet := make(map[string]struct{})
	for _, a := range assignments {
		prIDsSet[a.PullRequestID] = struct{}{}
	}
	uniquePRIDs := make([]string, 0, len(prIDsSet))
	for id := range prIDsSet {
		uniquePRIDs = append(uniquePRIDs, id)
	}

	currentReviewersMap, err := s.prRepo.GetReviewersByPRs(ctx, tx, uniquePRIDs)
	if err != nil {
		return err
	}

	replacements := []domain.ReviewReplacement{}
	removals := []domain.ReviewAssignment{}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, assignment := range assignments {
		prID := assignment.PullRequestID
		currentReviewers := currentReviewersMap[prID]

		var candidate *domain.User

		if len(activeUsers) > 0 {
			startIdx := rng.Intn(len(activeUsers))
			for i := 0; i < len(activeUsers); i++ {
				idx := (startIdx + i) % len(activeUsers)
				u := activeUsers[idx]

				// Constraints
				if u.ID == assignment.AuthorID {
					continue
				}
				alreadyReviewer := false
				for _, rID := range currentReviewers {
					if rID == u.ID {
						alreadyReviewer = true
						break
					}
				}
				if alreadyReviewer {
					continue
				}

				candidate = &u
				break
			}
		}

		if candidate != nil {
			replacements = append(replacements, domain.ReviewReplacement{
				PullRequestID: prID,
				OldUserID:     assignment.ReviewerID,
				NewUserID:     candidate.ID,
			})
			currentReviewersMap[prID] = append(currentReviewersMap[prID], candidate.ID)
		} else {
			removals = append(removals, assignment)
		}
	}

	if len(replacements) > 0 {
		if err := s.prRepo.ReplaceReviewersBulk(ctx, tx, replacements); err != nil {
			return err
		}
	}
	if len(removals) > 0 {
		if err := s.prRepo.RemoveReviewersBulk(ctx, tx, removals); err != nil {
			return err
		}
	}

	return tx.Commit()
}
