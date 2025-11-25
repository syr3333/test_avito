package handlers

import (
	"encoding/json"
	"net/http"

	"avito/internal/domain"
	"avito/internal/dto"
	"avito/internal/service"
)

// TeamHandler handles team endpoints
type TeamHandler struct {
	teamService *service.TeamService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// AddTeam handles POST /team/add
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert DTO to domain and create team
	domainTeam := req.ToDomain()
	if err := h.teamService.CreateTeam(r.Context(), domainTeam); err != nil {
		WriteAppError(w, err)
		return
	}

	// Fetch created team to return
	team, err := h.teamService.GetTeam(r.Context(), req.Name)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO and return
	response := dto.TeamFromDomain(team)
	WriteJSON(w, http.StatusCreated, response)
}

// GetTeam handles GET /team/get
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		WriteError(w, http.StatusBadRequest, domain.ErrCodeInvalidRequest, "team_name is required")
		return
	}

	// Validate team name
	if err := dto.ValidateTeamName(teamName); err != nil {
		WriteAppError(w, err)
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	// Convert domain to DTO and return
	response := dto.TeamFromDomain(team)
	WriteJSON(w, http.StatusOK, response)
}
