package handlers

import (
	"encoding/json"
	"net/http"

	"avito/internal/domain"
	"avito/internal/dto"
	"avito/internal/service"
)

// PullRequestHandler handles PR endpoints
type PullRequestHandler struct {
	prService *service.PullRequestService
}

// NewPullRequestHandler creates a new PR handler
func NewPullRequestHandler(prService *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
	}
}

// CreatePR handles POST /pullRequest/create
func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req dto.PullRequestCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteAppError(w, err)
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.ID, req.Name, req.AuthorID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO and return
	response := dto.PRFromDomain(pr)
	WriteJSON(w, http.StatusCreated, PRResponse{PR: response})
}

// MergePR handles POST /pullRequest/merge
func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req dto.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteAppError(w, err)
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO and return
	response := dto.PRFromDomain(pr)
	WriteJSON(w, http.StatusOK, PRResponse{PR: response})
}

func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req dto.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteAppError(w, err)
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)

	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO and return
	response := dto.PRFromDomain(pr)
	WriteJSON(w, http.StatusOK, PRReassignResponse{PR: response, ReplacedBy: replacedBy})
}
