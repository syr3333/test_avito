package integration

import (
	"context"
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

func TestTeamIntegration_MassDeactivate(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.Server.Close()

	// 1. Create Team with Author and 4 potential reviewers
	teamName := "devops"
	authorID := "author_devops"
	users := []string{authorID, "u1", "u2", "u3", "u4"}
	createTeamWithUsers(t, env.BaseURL(), teamName, users...)

	// 2. Create PR
	prID := "pr_mass_deactivate"
	createPR(t, env.BaseURL(), prID, "Infrastructure Update", authorID)

	// 3. Find current reviewers using Repo
	ctx := context.Background()
	pr, err := env.PRRepo.Get(ctx, prID)
	require.NoError(t, err)
	require.NotEmpty(t, pr.AssignedReviewers, "PR should have reviewers assigned")

	initialReviewers := pr.AssignedReviewers
	t.Logf("Initial reviewers: %v", initialReviewers)

	// 4. Select reviewers to deactivate (all currently assigned)
	toDeactivate := initialReviewers

	// 5. Mass deactivate them
	reqBody := dto.MassDeactivateRequest{
		TeamName: teamName,
		UserIDs:  toDeactivate,
	}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodPost,
		Path:   "/team/users/deactivate",
		Body:   reqBody,
	})
	assertStatusCode(t, resp, http.StatusOK)

	// 6. Verify Users are deactivated
	for _, uid := range toDeactivate {
		u, err := env.UserRepo.Get(ctx, uid)
		require.NoError(t, err)
		assert.False(t, u.IsActive, "User %s should be inactive", uid)
	}

	// 7. Verify PR reviewers are updated
	prUpdated, err := env.PRRepo.Get(ctx, prID)
	require.NoError(t, err)

	newReviewers := prUpdated.AssignedReviewers
	t.Logf("New reviewers: %v", newReviewers)

	// Assertions
	for _, r := range newReviewers {
		// New reviewers must NOT be in the deactivated list
		assert.NotContains(t, toDeactivate, r, "Deactivated user %s should not be a reviewer", r)
		// Should verify they are from u1..u4
		assert.Contains(t, users, r)
		assert.NotEqual(t, authorID, r)
	}

	// Since we had 4 eligible reviewers (u1-u4) and deactivated 2 (assuming standard 2 assigned),
	// there should be 2 active candidates left (u3, u4).
	// The logic should have replaced them.
	assert.Len(t, newReviewers, len(initialReviewers), "Should preserve number of reviewers if candidates exist")

	// Verify uniqueness
	unique := make(map[string]bool)
	for _, r := range newReviewers {
		unique[r] = true
	}
	assert.Equal(t, len(newReviewers), len(unique), "Reviewers should be unique")
}

func TestTeamIntegration_MassDeactivate_NoReplacement(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.Server.Close()

	// 1. Create Team with Author and 1 reviewer
	teamName := "small_team"
	authorID := "author_small"
	reviewerID := "only_reviewer"
	createTeamWithUsers(t, env.BaseURL(), teamName, authorID, reviewerID)

	// 2. Create PR
	prID := "pr_no_replacement"
	createPR(t, env.BaseURL(), prID, "Small Update", authorID)

	// 3. Check assignment
	ctx := context.Background()
	pr, err := env.PRRepo.Get(ctx, prID)
	require.NoError(t, err)
	require.Contains(t, pr.AssignedReviewers, reviewerID)

	// 4. Deactivate the only reviewer
	reqBody := dto.MassDeactivateRequest{
		TeamName: teamName,
		UserIDs:  []string{reviewerID},
	}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodPost,
		Path:   "/team/users/deactivate",
		Body:   reqBody,
	})
	assertStatusCode(t, resp, http.StatusOK)

	// 5. Verify PR has no reviewers (assignment removed)
	prUpdated, err := env.PRRepo.Get(ctx, prID)
	require.NoError(t, err)
	assert.Empty(t, prUpdated.AssignedReviewers, "Should remove assignment if no candidates available")
}
