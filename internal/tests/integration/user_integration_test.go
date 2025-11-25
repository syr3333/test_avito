package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"avito/internal/domain"
)

func TestUserIntegration_SetActiveDeactivate(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "alice", "bob")
	user := setUserActive(t, env.BaseURL(), "alice", false)
	assert.Equal(t, "alice", user.ID)
	assert.False(t, user.IsActive)
	assert.Equal(t, "backend", user.TeamName)
}

func TestUserIntegration_SetActiveActivate(t *testing.T) {
	env := setupTestEnvironment(t)
	members := []domain.TeamMember{
		{UserID: "charlie", Username: "Charlie", IsActive: false},
	}
	createTeam(t, env.BaseURL(), "payments", members)
	user := setUserActive(t, env.BaseURL(), "charlie", true)
	assert.True(t, user.IsActive)
}

func TestUserIntegration_SetActiveNonexistentUser(t *testing.T) {
	env := setupTestEnvironment(t)
	reqBody := map[string]interface{}{"user_id": "nonexistent", "is_active": false}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/users/setIsActive", Body: reqBody})
	assertStatusCode(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "NOT_FOUND")
}

func TestUserIntegration_GetReviewEmptyList(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "alice", "bob")
	review := getUserReview(t, env.BaseURL(), "alice")
	assert.Equal(t, "alice", review.UserID)
	assert.NotNil(t, review.PullRequests)
	assert.Len(t, review.PullRequests, 0)
}

func TestUserIntegration_GetReviewWithPRs(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "reviewer1", "reviewer2")
	createPR(t, env.BaseURL(), "pr-1", "Feature 1", "author")
	review := getUserReview(t, env.BaseURL(), "reviewer1")
	assert.Equal(t, "reviewer1", review.UserID)
	assert.Len(t, review.PullRequests, 1)
	assert.Equal(t, "pr-1", review.PullRequests[0].ID)
	assert.Equal(t, "Feature 1", review.PullRequests[0].Name)
	assert.Equal(t, "OPEN", review.PullRequests[0].Status)
}

func TestUserIntegration_InactiveUserNotAssigned(t *testing.T) {
	env := setupTestEnvironment(t)
	members := []domain.TeamMember{
		{UserID: "author", Username: "Author", IsActive: true},
		{UserID: "inactive_rev", Username: "InactiveRev", IsActive: false},
	}
	createTeam(t, env.BaseURL(), "team1", members)
	pr := createPR(t, env.BaseURL(), "pr-inactive", "Test", "author")
	assert.Len(t, pr.AssignedReviewers, 0)
}

func TestUserIntegration_GetReviewUserNotFound(t *testing.T) {
	env := setupTestEnvironment(t)
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodGet,
		Path:   "/users/getReview?user_id=nonexistent-user",
	})
	assertStatusCode(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "NOT_FOUND")
}
