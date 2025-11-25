package dto

import (
	"strings"
	"time"

	"avito/internal/domain"
)

type PullRequestCreateRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

func (r *PullRequestCreateRequest) Validate() error {
	if err := ValidatePullRequestID(r.ID); err != nil {
		return err
	}
	if err := ValidatePullRequestName(r.Name); err != nil {
		return err
	}
	if err := ValidateUserID(r.AuthorID); err != nil {
		return err
	}
	return nil
}

// MergePRRequest - запрос на мерж PR
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func (r *MergePRRequest) Validate() error {
	return ValidatePullRequestID(r.PullRequestID)
}

// ReassignReviewerRequest - запрос на переназначение ревьюера
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_user_id"`
}

// Validate проверяет корректность данных в запросе
func (r *ReassignReviewerRequest) Validate() error {
	if err := ValidatePullRequestID(r.PullRequestID); err != nil {
		return err
	}
	if err := ValidateUserID(r.OldReviewerID); err != nil {
		return err
	}
	return nil
}

// PullRequestResponse - ответ с данными PR
type PullRequestResponse struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

// ToDomain преобразует DTO в domain модель
func (r *PullRequestCreateRequest) ToDomain() *domain.PullRequest {
	return &domain.PullRequest{
		ID:                r.ID,
		Name:              r.Name,
		AuthorID:          r.AuthorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{},
		CreatedAt:         time.Now(),
	}
}

// PRFromDomain преобразует domain модель в DTO
func PRFromDomain(pr *domain.PullRequest) PullRequestResponse {
	return PullRequestResponse{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func ValidatePullRequestID(id string) error {
	if strings.TrimSpace(id) == "" {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "pull_request_id cannot be empty")
	}
	if len(id) > 255 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "pull_request_id too long (max 255 characters)")
	}
	if !idRegex.MatchString(id) {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "pull_request_id contains invalid characters")
	}
	return nil
}

func ValidatePullRequestName(name string) error {
	if strings.TrimSpace(name) == "" {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "pull_request_name cannot be empty")
	}
	if len(name) > 500 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "pull_request_name too long (max 500 characters)")
	}
	return nil
}
