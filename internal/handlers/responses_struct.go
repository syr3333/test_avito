package handlers

import (
	"avito/internal/dto"
)

type PRResponse struct {
	PR dto.PullRequestResponse `json:"pr"`
}

type PRReassignResponse struct {
	PR         dto.PullRequestResponse `json:"pr"`
	ReplacedBy string                  `json:"replaced_by"`
}

type UserReviewResponse struct {
	UserID       string                 `json:"user_id"`
	PullRequests []dto.PullRequestShort `json:"pull_requests"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
