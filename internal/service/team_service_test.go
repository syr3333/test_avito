package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"avito/internal/domain"
)

type mockTeamRepo struct {
	existsFn func(ctx context.Context, teamName string) (bool, error)
	createFn func(ctx context.Context, tx *sql.Tx, teamName string) error
	getFn    func(ctx context.Context, teamName string) (*domain.Team, error)
}

func (m *mockTeamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, teamName)
	}
	return false, nil
}

func (m *mockTeamRepo) Create(ctx context.Context, tx *sql.Tx, teamName string) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, teamName)
	}
	return nil
}

func (m *mockTeamRepo) Get(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.getFn != nil {
		return m.getFn(ctx, teamName)
	}
	return nil, sql.ErrNoRows
}

type mockUserRepo struct {
	createFn func(ctx context.Context, tx *sql.Tx, user *domain.User) error
	getFn    func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, user)
	}
	return nil
}

func (m *mockUserRepo) Get(ctx context.Context, userID string) (*domain.User, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID)
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *mockUserRepo) GetByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetActiveTeammates(ctx context.Context, authorID string, limit int) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindReplacementReviewer(ctx context.Context, tx *sql.Tx, teamName string, excludeIDs []string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) SetActive(ctx context.Context, userID string, isActive bool) error {
	return nil
}

type mockTxManager struct {
	beginFn func(ctx context.Context) (*sql.Tx, error)
}

func (m *mockTxManager) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if m.beginFn != nil {
		return m.beginFn(ctx)
	}

	return &sql.Tx{}, nil
}

func TestTeamService_CreateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("team already exists", func(t *testing.T) {
		teamRepo := &mockTeamRepo{
			existsFn: func(ctx context.Context, teamName string) (bool, error) {
				return true, nil
			},
		}
		userRepo := &mockUserRepo{}
		txMgr := &mockTxManager{}

		service := NewTeamService(teamRepo, userRepo, txMgr)

		team := &domain.Team{
			Name: "backend",
			Members: []domain.TeamMember{
				{UserID: "u1", Username: "Alice", IsActive: true},
			},
		}

		err := service.CreateTeam(ctx, team)

		if err == nil {
			t.Error("Expected error for existing team")
		}

		appErr, ok := err.(*domain.AppError)
		if !ok || appErr.Code != domain.ErrCodeTeamExists {
			t.Errorf("Expected TEAM_EXISTS error, got %v", err)
		}
	})

	t.Run("repository error on exists check", func(t *testing.T) {
		expectedErr := errors.New("database error")
		teamRepo := &mockTeamRepo{
			existsFn: func(ctx context.Context, teamName string) (bool, error) {
				return false, expectedErr
			},
		}
		userRepo := &mockUserRepo{}
		txMgr := &mockTxManager{}

		service := NewTeamService(teamRepo, userRepo, txMgr)

		team := &domain.Team{
			Name: "backend",
			Members: []domain.TeamMember{
				{UserID: "u1", Username: "Alice", IsActive: true},
			},
		}

		err := service.CreateTeam(ctx, team)

		if err != expectedErr {
			t.Errorf("Expected database error, got %v", err)
		}
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedTeam := &domain.Team{
			Name: "backend",
			Members: []domain.TeamMember{
				{UserID: "u1", Username: "Alice", IsActive: true},
			},
		}

		teamRepo := &mockTeamRepo{
			getFn: func(ctx context.Context, teamName string) (*domain.Team, error) {
				return expectedTeam, nil
			},
		}
		userRepo := &mockUserRepo{}
		txMgr := &mockTxManager{}

		service := NewTeamService(teamRepo, userRepo, txMgr)
		team, err := service.GetTeam(ctx, "backend")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if team.Name != "backend" {
			t.Errorf("Expected team name 'backend', got '%s'", team.Name)
		}
	})

	t.Run("team not found", func(t *testing.T) {
		teamRepo := &mockTeamRepo{
			getFn: func(ctx context.Context, teamName string) (*domain.Team, error) {
				return nil, domain.NewAppError(domain.ErrCodeNotFound, "team not found")
			},
		}
		userRepo := &mockUserRepo{}
		txMgr := &mockTxManager{}

		service := NewTeamService(teamRepo, userRepo, txMgr)
		_, err := service.GetTeam(ctx, "nonexistent")

		if err == nil {
			t.Error("Expected error for nonexistent team")
		}

		appErr, ok := err.(*domain.AppError)
		if !ok || appErr.Code != domain.ErrCodeNotFound {
			t.Errorf("Expected NOT_FOUND error, got %v", err)
		}
	})
}
