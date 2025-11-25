package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avito/internal/domain"
	"avito/internal/dto"
)

func TestTeamIntegration_CreateTeam(t *testing.T) {
	env := setupTestEnvironment(t)
	members := []domain.TeamMember{
		{UserID: "alice", Username: "Alice", IsActive: true},
		{UserID: "bob", Username: "Bob", IsActive: true},
	}
	team := createTeam(t, env.BaseURL(), "backend", members)
	assert.Equal(t, "backend", team.TeamName)
	assert.Len(t, team.Members, 2)
	assert.Equal(t, "alice", team.Members[0].UserID)
	assert.Equal(t, "bob", team.Members[1].UserID)
}

func TestTeamIntegration_CreateDuplicateTeam(t *testing.T) {
	env := setupTestEnvironment(t)
	members := []domain.TeamMember{
		{UserID: "alice", Username: "Alice", IsActive: true},
	}
	createTeam(t, env.BaseURL(), "backend", members)

	// Convert to DTO members
	dtoMembers := []dto.TeamMember{
		{UserID: "alice", Username: "Alice", IsActive: true},
	}
	reqBody := dto.TeamRequest{Name: "backend", Members: dtoMembers}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/team/add", Body: reqBody})
	assertStatusCode(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "TEAM_EXISTS")
}

func TestTeamIntegration_CreateTeamWithEmptyMembers(t *testing.T) {
	env := setupTestEnvironment(t)
	reqBody := dto.TeamRequest{Name: "empty_team", Members: []dto.TeamMember{}}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/team/add", Body: reqBody})
	assertStatusCode(t, resp, http.StatusInternalServerError)
	assertErrorCode(t, resp, "INVALID_INPUT")
}

func TestTeamIntegration_GetTeam(t *testing.T) {
	env := setupTestEnvironment(t)
	members := []domain.TeamMember{
		{UserID: "charlie", Username: "Charlie", IsActive: true},
	}
	createTeam(t, env.BaseURL(), "payments", members)
	team, resp := getTeam(t, env.BaseURL(), "payments")
	assertStatusCode(t, resp, http.StatusOK)
	require.NotNil(t, team)
	assert.Equal(t, "payments", team.TeamName)
	assert.Len(t, team.Members, 1)
}

func TestTeamIntegration_GetNonexistentTeam(t *testing.T) {
	env := setupTestEnvironment(t)
	_, resp := getTeam(t, env.BaseURL(), "nonexistent")
	assertStatusCode(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "NOT_FOUND")
}

func TestTeamIntegration_UpsertBehavior(t *testing.T) {
	env := setupTestEnvironment(t)
	members1 := []domain.TeamMember{
		{UserID: "user1", Username: "User1", IsActive: true},
	}
	createTeam(t, env.BaseURL(), "test_team", members1)
	members2 := []dto.TeamMember{
		{UserID: "user1", Username: "UpdatedUser1", IsActive: false},
		{UserID: "user2", Username: "User2", IsActive: true},
	}
	reqBody := dto.TeamRequest{Name: "test_team", Members: members2}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/team/add", Body: reqBody})
	assertStatusCode(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "TEAM_EXISTS")
}

func TestTeamIntegration_ValidateTeamName(t *testing.T) {
	env := setupTestEnvironment(t)
	tests := []struct {
		name         string
		teamName     string
		expectStatus int
	}{
		{"empty name", "", http.StatusInternalServerError},
		{"valid name", "valid_team", http.StatusCreated},
		{"name with spaces", "Team Name", http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members := []dto.TeamMember{{UserID: "u1", Username: "U1", IsActive: true}}
			reqBody := dto.TeamRequest{Name: tt.teamName, Members: members}
			resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/team/add", Body: reqBody})
			assertStatusCode(t, resp, tt.expectStatus)
		})
	}
}
