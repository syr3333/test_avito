package repository

import (
	"context"
	"database/sql"

	"avito/internal/domain"

	sq "github.com/Masterminds/squirrel"
)

type prRepo struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewPullRequestRepository(db *sql.DB) PullRequestRepository {
	return &prRepo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *prRepo) Create(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error {
	query, args, err := r.builder.
		Insert("pull_requests").
		Columns("id", "name", "author_id", "status", "created_at").
		Values(pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt).
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

func (r *prRepo) Get(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query, args, err := r.builder.
		Select("id", "name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where(sq.Eq{"id": prID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var pr domain.PullRequest
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewAppError(domain.ErrCodeNotFound, "PR not found")
	}
	if err != nil {
		return nil, err
	}

	// reviewers
	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *prRepo) GetForUpdate(ctx context.Context, tx *sql.Tx, prID string) (*domain.PullRequest, error) {
	query, args, err := r.builder.
		Select("id", "name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where(sq.Eq{"id": prID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, err
	}

	var pr domain.PullRequest

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, args...).Scan(
			&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, args...).Scan(
			&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
		)
	}

	if err == sql.ErrNoRows {
		return nil, domain.NewAppError(domain.ErrCodeNotFound, "PR not found")
	}
	if err != nil {
		return nil, err
	}

	// Get reviewers
	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *prRepo) Update(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error {
	query, args, err := r.builder.
		Update("pull_requests").
		Set("name", pr.Name).
		Set("author_id", pr.AuthorID).
		Set("status", pr.Status).
		Set("merged_at", pr.MergedAt).
		Where(sq.Eq{"id": pr.ID}).
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

func (r *prRepo) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	query, args, err := r.builder.
		Select("1").
		Prefix("SELECT EXISTS(").
		From("pull_requests").
		Where(sq.Eq{"id": prID}).
		Suffix(")").
		ToSql()
	if err != nil {
		return false, err
	}
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (r *prRepo) GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	query, args, err := r.builder.
		Select("pr.id", "pr.name", "pr.author_id", "pr.status").
		From("pull_requests pr").
		Join("pr_reviewers prr ON pr.id = prr.pull_request_id").
		Where(sq.Eq{"prr.user_id": userID}).
		OrderBy("pr.created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prs := make([]domain.PullRequestShort, 0)
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

func (r *prRepo) AddReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error {
	query, args, err := r.builder.
		Insert("pr_reviewers").
		Columns("pull_request_id", "user_id").
		Values(prID, userID).
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

func (r *prRepo) RemoveReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error {
	query, args, err := r.builder.
		Delete("pr_reviewers").
		Where(sq.Eq{"pull_request_id": prID, "user_id": userID}).
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

func (r *prRepo) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	query, args, err := r.builder.
		Select("user_id").
		From("pr_reviewers").
		Where(sq.Eq{"pull_request_id": prID}).
		OrderBy("assigned_at").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewers := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, userID)
	}

	return reviewers, rows.Err()
}
