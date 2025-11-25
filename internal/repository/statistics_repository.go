package repository

import (
	"context"
	"database/sql"

	"avito/internal/domain"

	sq "github.com/Masterminds/squirrel"
)

type StatisticsRepository interface {
	GetAssignmentsByUser(ctx context.Context) ([]domain.AssignmentStat, error)
	GetAssignmentsByPR(ctx context.Context) ([]domain.AssignmentStat, error)
	GetTotalPRs(ctx context.Context) (int, error)
	GetActiveUsersCount(ctx context.Context) (int, error)
	GetTeamsCount(ctx context.Context) (int, error)
}

type statisticsRepository struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewStatisticsRepository(db *sql.DB) StatisticsRepository {
	return &statisticsRepository{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *statisticsRepository) GetAssignmentsByUser(ctx context.Context) ([]domain.AssignmentStat, error) {
	query, args, err := r.builder.
		Select("user_id", "COUNT(*) as count").
		From("pr_reviewers").
		GroupBy("user_id").
		OrderBy("count DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.AssignmentStat
	for rows.Next() {
		var stat domain.AssignmentStat
		if err := rows.Scan(&stat.ID, &stat.Count); err != nil {
			return nil, err
		}
		result = append(result, stat)
	}

	return result, rows.Err()
}

func (r *statisticsRepository) GetAssignmentsByPR(ctx context.Context) ([]domain.AssignmentStat, error) {
	query, args, err := r.builder.
		Select("pull_request_id", "COUNT(*) as count").
		From("pr_reviewers").
		GroupBy("pull_request_id").
		OrderBy("count DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.AssignmentStat
	for rows.Next() {
		var stat domain.AssignmentStat
		if err := rows.Scan(&stat.ID, &stat.Count); err != nil {
			return nil, err
		}
		result = append(result, stat)
	}

	return result, rows.Err()
}

func (r *statisticsRepository) GetTotalPRs(ctx context.Context) (int, error) {
	var count int
	query, args, err := r.builder.
		Select("COUNT(*)").
		From("pull_requests").
		ToSql()
	if err != nil {
		return 0, err
	}
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *statisticsRepository) GetActiveUsersCount(ctx context.Context) (int, error) {
	var count int
	query, args, err := r.builder.
		Select("COUNT(*)").
		From("users").
		Where(sq.Eq{"is_active": true}).
		ToSql()
	if err != nil {
		return 0, err
	}
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// GetTeamsCount retrieves total count of teams
func (r *statisticsRepository) GetTeamsCount(ctx context.Context) (int, error) {
	var count int
	query, args, err := r.builder.
		Select("COUNT(*)").
		From("teams").
		ToSql()
	if err != nil {
		return 0, err
	}
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}
