package repository

import (
	"context"
	"database/sql"

	"avito/internal/domain"

	sq "github.com/Masterminds/squirrel"
)

type teamRepo struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewTeamRepository(db *sql.DB) TeamRepository {
	return &teamRepo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *teamRepo) Create(ctx context.Context, tx *sql.Tx, teamName string) error {
	query, args, err := r.builder.
		Insert("teams").
		Columns("name").
		Values(teamName).
		ToSql()
	if err != nil {
		return err
	}

	if tx != nil {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = r.db.ExecContext(ctx, query, args...)
	}

	return err
}

func (r *teamRepo) Get(ctx context.Context, teamName string) (*domain.Team, error) {
	// Get team members
	query, args, err := r.builder.
		Select("u.id", "u.username", "u.is_active").
		From("users u").
		Where(sq.Eq{"u.team_name": teamName}).
		OrderBy("u.username").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		var exists bool
		q, a, err := r.builder.
			Select("1").
			Prefix("SELECT EXISTS(").
			From("teams").
			Where(sq.Eq{"name": teamName}).
			Suffix(")").
			ToSql()
		if err != nil {
			return nil, err
		}
		err = r.db.QueryRowContext(ctx, q, a...).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.NewAppError(domain.ErrCodeNotFound, "team not found")
		}
	}

	return &domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (r *teamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	query, args, err := r.builder.
		Select("1").
		Prefix("SELECT EXISTS(").
		From("teams").
		Where(sq.Eq{"name": teamName}).
		Suffix(")").
		ToSql()
	if err != nil {
		return false, err
	}
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}
