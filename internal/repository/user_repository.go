package repository

import (
	"context"
	"database/sql"

	"avito/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

type userRepo struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *userRepo) Create(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	query, args, err := r.builder.
		Insert("users").
		Columns("id", "username", "team_name", "is_active").
		Values(user.ID, user.Username, user.TeamName, user.IsActive).
		Suffix("ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active").
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

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query, args, err := r.builder.
		Update("users").
		Set("username", user.Username).
		Set("team_name", user.TeamName).
		Set("is_active", user.IsActive).
		Where(sq.Eq{"id": user.ID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *userRepo) Get(ctx context.Context, userID string) (*domain.User, error) {
	query, args, err := r.builder.
		Select("id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.TeamName, &user.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewAppError(domain.ErrCodeNotFound, "user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	query, args, err := r.builder.
		Select("id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"team_name": teamName}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *userRepo) GetActiveTeammates(ctx context.Context, authorID string, limit int) ([]domain.User, error) {
	// Squirrel рекомендуется к использования, хоть и не поддерживает подзапрос, поэтому пишем сырой SQL
	query := `
		SELECT u.id, u.username, u.team_name, u.is_active
		FROM users u
		WHERE u.team_name = (SELECT team_name FROM users WHERE id = $1)
		  AND u.is_active = true
		  AND u.id != $1
		ORDER BY RANDOM()
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, authorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *userRepo) FindReplacementReviewer(ctx context.Context, tx *sql.Tx, teamName string, excludeIDs []string) (*domain.User, error) {
	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		  AND is_active = true
		  AND id != ALL($2)
		ORDER BY RANDOM()
		LIMIT 1
	`

	var user domain.User
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, teamName, pq.Array(excludeIDs)).Scan(
			&user.ID, &user.Username, &user.TeamName, &user.IsActive,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, teamName, pq.Array(excludeIDs)).Scan(
			&user.ID, &user.Username, &user.TeamName, &user.IsActive,
		)
	}

	if err == sql.ErrNoRows {
		return nil, domain.NewAppError(domain.ErrCodeNoCandidate, "no active replacement candidate in team")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) SetActive(ctx context.Context, userID string, isActive bool) error {
	query, args, err := r.builder.
		Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *userRepo) DeactivateMany(ctx context.Context, tx *sql.Tx, userIDs []string) error {
	query, args, err := r.builder.
		Update("users").
		Set("is_active", false).
		Where(sq.Eq{"id": userIDs}).
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

func (r *userRepo) GetActiveUsersByTeam(ctx context.Context, tx *sql.Tx, teamName string) ([]domain.User, error) {
	query, args, err := r.builder.
		Select("id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"team_name": teamName, "is_active": true}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	if tx != nil {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = r.db.QueryContext(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}
