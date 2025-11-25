package repository

import (
	"context"
	"database/sql"

	"avito/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, tx *sql.Tx, teamName string) error
	Get(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

type UserRepository interface {
	Create(ctx context.Context, tx *sql.Tx, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, userID string) (*domain.User, error)
	GetByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	GetActiveTeammates(ctx context.Context, authorID string, limit int) ([]domain.User, error)
	FindReplacementReviewer(ctx context.Context, tx *sql.Tx, teamName string, excludeIDs []string) (*domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) error
	DeactivateMany(ctx context.Context, tx *sql.Tx, userIDs []string) error
	GetActiveUsersByTeam(ctx context.Context, tx *sql.Tx, teamName string) ([]domain.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error
	Get(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetForUpdate(ctx context.Context, tx *sql.Tx, prID string) (*domain.PullRequest, error)
	Update(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error
	Exists(ctx context.Context, prID string) (bool, error)
	GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	AddReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error
	RemoveReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	GetOpenAssignmentsByReviewers(ctx context.Context, tx *sql.Tx, reviewerIDs []string) ([]domain.ReviewAssignment, error)
	ReplaceReviewersBulk(ctx context.Context, tx *sql.Tx, replacements []domain.ReviewReplacement) error
	GetReviewersByPRs(ctx context.Context, tx *sql.Tx, prIDs []string) (map[string][]string, error)
	RemoveReviewersBulk(ctx context.Context, tx *sql.Tx, assignments []domain.ReviewAssignment) error
}

type TransactionManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
