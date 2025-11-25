package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avito/internal/domain"
	"avito/internal/dto"
)

type HTTPRequest struct {
	Method string
	Path   string
	Body   interface{}
}

func doRequest(t *testing.T, baseURL string, req HTTPRequest) *http.Response {
	var body io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		require.NoError(t, err)
		body = bytes.NewBuffer(jsonData)
	}

	httpReq, err := http.NewRequest(req.Method, baseURL+req.Path, body)
	require.NoError(t, err)

	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)

	return resp
}

func parseJSON(t *testing.T, resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}

func assertStatusCode(t *testing.T, resp *http.Response, expectedStatus int) {
	if resp.StatusCode != expectedStatus {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Response body: %s", string(bodyBytes))
	}
	assert.Equal(t, expectedStatus, resp.StatusCode)
}

func assertErrorCode(t *testing.T, resp *http.Response, expectedCode string) {
	var errorResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	parseJSON(t, resp, &errorResp)
	assert.Equal(t, expectedCode, errorResp.Error.Code)
}

func createTeam(t *testing.T, baseURL, teamName string, members []domain.TeamMember) *dto.TeamResponse {
	// Convert domain.TeamMember to dto.TeamMember
	dtoMembers := make([]dto.TeamMember, len(members))
	for i, m := range members {
		dtoMembers[i] = dto.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	reqBody := dto.TeamRequest{Name: teamName, Members: dtoMembers}

	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodPost, Path: "/team/add", Body: reqBody})

	assertStatusCode(t, resp, http.StatusCreated)

	var teamResp dto.TeamResponse
	parseJSON(t, resp, &teamResp)

	return &teamResp
}

func getTeam(t *testing.T, baseURL, teamName string) (*dto.TeamResponse, *http.Response) {
	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodGet, Path: fmt.Sprintf("/team/get?team_name=%s", teamName)})

	if resp.StatusCode != http.StatusOK {
		return nil, resp
	}

	var team dto.TeamResponse
	parseJSON(t, resp, &team)
	return &team, resp
}

type UserSetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func setUserActive(t *testing.T, baseURL, userID string, isActive bool) *dto.UserResponse {
	reqBody := UserSetActiveRequest{UserID: userID, IsActive: isActive}

	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodPost, Path: "/users/setIsActive", Body: reqBody})

	assertStatusCode(t, resp, http.StatusOK)

	var userResp dto.UserResponse
	parseJSON(t, resp, &userResp)

	return &userResp
}

type GetReviewResponse struct {
	UserID       string                 `json:"user_id"`
	PullRequests []dto.PullRequestShort `json:"pull_requests"`
}

func getUserReview(t *testing.T, baseURL, userID string) *GetReviewResponse {
	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodGet, Path: fmt.Sprintf("/users/getReview?user_id=%s", userID)})

	assertStatusCode(t, resp, http.StatusOK)

	var reviewResp GetReviewResponse
	parseJSON(t, resp, &reviewResp)

	return &reviewResp
}

func createPR(t *testing.T, baseURL, prID, prName, authorID string) *dto.PullRequestResponse {
	reqBody := dto.PullRequestCreateRequest{ID: prID, Name: prName, AuthorID: authorID}

	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/create", Body: reqBody})

	assertStatusCode(t, resp, http.StatusCreated)

	var prResp struct {
		PR dto.PullRequestResponse `json:"pr"`
	}
	parseJSON(t, resp, &prResp)

	return &prResp.PR
}

func mergePR(t *testing.T, baseURL, prID string) *dto.PullRequestResponse {
	reqBody := map[string]string{"pull_request_id": prID}

	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/merge", Body: reqBody})

	assertStatusCode(t, resp, http.StatusOK)

	var prResp struct {
		PR dto.PullRequestResponse `json:"pr"`
	}
	parseJSON(t, resp, &prResp)

	return &prResp.PR
}

type ReassignResponse struct {
	PR         dto.PullRequestResponse `json:"pr"`
	ReplacedBy string                  `json:"replaced_by"`
}

func reassignReviewer(t *testing.T, baseURL, prID, oldUserID string) *ReassignResponse {
	reqBody := map[string]string{"pull_request_id": prID, "old_user_id": oldUserID}

	resp := doRequest(t, baseURL, HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/reassign", Body: reqBody})

	assertStatusCode(t, resp, http.StatusOK)

	var reassignResp ReassignResponse
	parseJSON(t, resp, &reassignResp)

	return &reassignResp
}

func createTeamWithUsers(t *testing.T, baseURL, teamName string, userIDs ...string) *dto.TeamResponse {
	members := make([]domain.TeamMember, len(userIDs))
	for i, userID := range userIDs {
		members[i] = domain.TeamMember{UserID: userID, Username: fmt.Sprintf("User_%s", userID), IsActive: true}
	}
	return createTeam(t, baseURL, teamName, members)
}

func addUserToTeam(t *testing.T, baseURL, teamName, userID string, isActive bool) {
	t.Helper()

	// Get current team
	resp := doRequest(t, baseURL, HTTPRequest{
		Method: http.MethodGet,
		Path:   "/team/get?team_name=" + teamName,
	})
	assertStatusCode(t, resp, http.StatusOK)

	var currentTeam dto.TeamResponse
	parseJSON(t, resp, &currentTeam)

	// Add new member
	newMember := dto.TeamMember{
		UserID:   userID,
		Username: fmt.Sprintf("User_%s", userID),
		IsActive: isActive,
	}

	// Update team with new member (using upsert logic)
	allMembers := append(currentTeam.Members, newMember)

	reqBody := dto.TeamRequest{
		Name:    teamName,
		Members: allMembers,
	}

	resp = doRequest(t, baseURL, HTTPRequest{
		Method: http.MethodPost,
		Path:   "/team/add",
		Body:   reqBody,
	})

	// Should return 200 (upsert) since team exists
	assertStatusCode(t, resp, http.StatusOK)
}
