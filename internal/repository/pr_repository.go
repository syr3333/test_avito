package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"avito/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
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

func (r *prRepo) GetOpenAssignmentsByReviewers(ctx context.Context, tx *sql.Tx, reviewerIDs []string) ([]domain.ReviewAssignment, error) {
	query := `
		SELECT prr.pull_request_id, prr.user_id, pr.author_id
		FROM pr_reviewers prr
		JOIN pull_requests pr ON prr.pull_request_id = pr.id
		WHERE prr.user_id = ANY($1) AND pr.status = 'OPEN'
	`

	var rows *sql.Rows
	var err error

	if tx != nil {
		rows, err = tx.QueryContext(ctx, query, pq.Array(reviewerIDs))
	} else {
		rows, err = r.db.QueryContext(ctx, query, pq.Array(reviewerIDs))
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []domain.ReviewAssignment{}
	for rows.Next() {
		var a domain.ReviewAssignment
		if err := rows.Scan(&a.PullRequestID, &a.ReviewerID, &a.AuthorID); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

func (r *prRepo) ReplaceReviewersBulk(ctx context.Context, tx *sql.Tx, replacements []domain.ReviewReplacement) error {
	if len(replacements) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(replacements))
	valueArgs := make([]interface{}, 0, len(replacements)*3)

	for i, rep := range replacements {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, rep.PullRequestID, rep.OldUserID, rep.NewUserID)
	}

	query := fmt.Sprintf(`
		UPDATE pr_reviewers AS t 
		SET user_id = v.new_user_id 
		FROM (VALUES %s) AS v(pr_id, old_user_id, new_user_id) 
		WHERE t.pull_request_id = v.pr_id AND t.user_id = v.old_user_id
	`, strings.Join(valueStrings, ","))

	if tx != nil {
		_, err := tx.ExecContext(ctx, query, valueArgs...)
		return err
	}
	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	return err
}

func (r *prRepo) GetReviewersByPRs(ctx context.Context, tx *sql.Tx, prIDs []string) (map[string][]string, error) {
	if len(prIDs) == 0 {
		return map[string][]string{}, nil
	}

	query := `SELECT pull_request_id, user_id FROM pr_reviewers WHERE pull_request_id = ANY($1)`

	var rows *sql.Rows
	var err error

	if tx != nil {
		rows, err = tx.QueryContext(ctx, query, pq.Array(prIDs))
	} else {
		rows, err = r.db.QueryContext(ctx, query, pq.Array(prIDs))
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var prID, userID string
		if err := rows.Scan(&prID, &userID); err != nil {
			return nil, err
		}
		result[prID] = append(result[prID], userID)
	}
	return result, rows.Err()
}

func (r *prRepo) RemoveReviewersBulk(ctx context.Context, tx *sql.Tx, assignments []domain.ReviewAssignment) error {
	if len(assignments) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(assignments))
	valueArgs := make([]interface{}, 0, len(assignments)*2)

	for i, a := range assignments {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, a.PullRequestID, a.ReviewerID)
	}

	query := fmt.Sprintf(`
		DELETE FROM pr_reviewers AS t 
		USING (VALUES %s) AS v(pr_id, user_id) 
		WHERE t.pull_request_id = v.pr_id AND t.user_id = v.user_id
	`, strings.Join(valueStrings, ","))

	if tx != nil {
		_, err := tx.ExecContext(ctx, query, valueArgs...)
		return err
	}
	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	return err
}
