package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatisticsIntegration_EmptyDatabase(t *testing.T) {
	env := setupTestEnvironment(t)
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: "GET", Path: "/statistics"})
	assertStatusCode(t, resp, 200)
	var stats map[string]interface{}
	parseJSON(t, resp, &stats)
	assert.Equal(t, float64(0), stats["total_prs"])
	assert.Equal(t, float64(0), stats["total_assignments"])
	assert.Equal(t, float64(0), stats["active_users"])
	assert.Equal(t, float64(0), stats["teams"])
}

func TestStatisticsIntegration_WithData(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2")
	createTeamWithUsers(t, env.BaseURL(), "frontend", "fe_author", "fe_rev")
	createPR(t, env.BaseURL(), "pr-1", "Feature 1", "author")
	createPR(t, env.BaseURL(), "pr-2", "Feature 2", "fe_author")
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: "GET", Path: "/statistics"})
	assertStatusCode(t, resp, 200)
	var stats map[string]interface{}
	parseJSON(t, resp, &stats)
	assert.Equal(t, float64(2), stats["total_prs"])
	assert.Equal(t, float64(3), stats["total_assignments"])
	assert.Equal(t, float64(5), stats["active_users"])
	assert.Equal(t, float64(2), stats["teams"])
}

func TestStatisticsIntegration_AssignmentsByUser(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2")
	createPR(t, env.BaseURL(), "pr-1", "Feature", "author")
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: "GET", Path: "/statistics"})
	assertStatusCode(t, resp, 200)
	var stats map[string]interface{}
	parseJSON(t, resp, &stats)
	assignmentsByUser := stats["assignments_by_user"].([]interface{})
	assert.Len(t, assignmentsByUser, 2)
	// Check that both reviewers have count = 1
	for _, item := range assignmentsByUser {
		stat := item.(map[string]interface{})
		assert.Contains(t, []string{"rev1", "rev2"}, stat["id"])
		assert.Equal(t, float64(1), stat["count"])
	}
}

func TestStatisticsIntegration_AssignmentsByPR(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2", "rev3")
	createPR(t, env.BaseURL(), "pr-single", "Feature", "author")
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: "GET", Path: "/statistics"})
	assertStatusCode(t, resp, 200)
	var stats map[string]interface{}
	parseJSON(t, resp, &stats)
	assignmentsByPR := stats["assignments_by_pr"].([]interface{})
	assert.Len(t, assignmentsByPR, 1)
	prStat := assignmentsByPR[0].(map[string]interface{})
	assert.Equal(t, "pr-single", prStat["id"])
	assert.Equal(t, float64(2), prStat["count"])
}
