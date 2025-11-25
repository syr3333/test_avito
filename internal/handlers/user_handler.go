package handlers

import (
	"encoding/json"
	"net/http"

	"avito/internal/domain"
	"avito/internal/dto"
	"avito/internal/repository"
	"avito/internal/service"
)

// UserHandler handles user endpoints
type UserHandler struct {
	userService *service.UserService
	prRepo      repository.PullRequestRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService, prRepo repository.PullRequestRepository) *UserHandler {
	return &UserHandler{
		userService: userService,
		prRepo:      prRepo,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req dto.SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteAppError(w, err)
		return
	}

	user, err := h.userService.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO and return
	response := dto.UserFromDomain(user)
	WriteJSON(w, http.StatusOK, response)
}

// GetReview handles GET /users/getReview
func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "user_id is required")
		return
	}

	// Validate user ID
	if err := dto.ValidateUserID(userID); err != nil {
		WriteAppError(w, err)
		return
	}

	prs, err := h.userService.GetReviewPRs(r.Context(), userID, h.prRepo)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO
	dtoPRs := dto.PullRequestsShortFromDomain(prs)

	WriteJSON(w, http.StatusOK, UserReviewResponse{
		UserID:       userID,
		PullRequests: dtoPRs,
	})
}
