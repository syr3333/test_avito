package service

import (
	"context"

	"avito/internal/domain"
	"avito/internal/repository"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	txMgr    repository.TransactionManager
}

func NewTeamService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	txMgr repository.TransactionManager,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
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
