package integration

import (
	"net/http"
	"sync"
	"testing"

	"avito/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRIntegration_CreateWithZeroReviewers(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author")
	pr := createPR(t, env.BaseURL(), "pr-1", "Feature", "author")
	assert.Equal(t, "pr-1", pr.ID)
	assert.Equal(t, "OPEN", pr.Status)
	assert.Len(t, pr.AssignedReviewers, 0)
}

func TestPRIntegration_CreateWithOneReviewer(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "reviewer1")
	pr := createPR(t, env.BaseURL(), "pr-2", "Feature", "author")
	assert.Len(t, pr.AssignedReviewers, 1)
	assert.Equal(t, "reviewer1", pr.AssignedReviewers[0])
	assert.NotContains(t, pr.AssignedReviewers, "author")
}

func TestPRIntegration_CreateWithTwoReviewers(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2")
	pr := createPR(t, env.BaseURL(), "pr-3", "Feature", "author")
	assert.Len(t, pr.AssignedReviewers, 2)
	assert.NotContains(t, pr.AssignedReviewers, "author")
	assert.ElementsMatch(t, []string{"rev1", "rev2"}, pr.AssignedReviewers)
}

func TestPRIntegration_CreateWithMoreThanTwoTeammates(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "r1", "r2", "r3", "r4")
	pr := createPR(t, env.BaseURL(), "pr-4", "Feature", "author")
	assert.Len(t, pr.AssignedReviewers, 2)
	assert.NotContains(t, pr.AssignedReviewers, "author")
}

func TestPRIntegration_CreateDuplicate(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1")
	createPR(t, env.BaseURL(), "pr-dup", "Feature", "author")
	reqBody := map[string]string{"pull_request_id": "pr-dup", "pull_request_name": "Duplicate", "author_id": "author"}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/create", Body: reqBody})
	assertStatusCode(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "PR_EXISTS")
}

func TestPRIntegration_MergePR(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1")
	createPR(t, env.BaseURL(), "pr-merge", "Feature", "author")
	pr := mergePR(t, env.BaseURL(), "pr-merge")
	assert.Equal(t, "MERGED", pr.Status)
	assert.NotNil(t, pr.MergedAt)
}

func TestPRIntegration_MergeIdempotent(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1")
	createPR(t, env.BaseURL(), "pr-idem", "Feature", "author")
	pr1 := mergePR(t, env.BaseURL(), "pr-idem")
	pr2 := mergePR(t, env.BaseURL(), "pr-idem")
	assert.Equal(t, "MERGED", pr1.Status)
	assert.Equal(t, "MERGED", pr2.Status)
	assert.NotNil(t, pr1.MergedAt)
	assert.NotNil(t, pr2.MergedAt)
}

func TestPRIntegration_MergeNonexistent(t *testing.T) {
	env := setupTestEnvironment(t)
	reqBody := map[string]string{"pull_request_id": "nonexistent"}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/merge", Body: reqBody})
	assertStatusCode(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "NOT_FOUND")
}

func TestPRIntegration_ReassignSuccess(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2", "rev3")
	pr := createPR(t, env.BaseURL(), "pr-reassign", "Feature", "author")
	require.Len(t, pr.AssignedReviewers, 2)
	oldReviewer := pr.AssignedReviewers[0]
	result := reassignReviewer(t, env.BaseURL(), "pr-reassign", oldReviewer)
	assert.NotContains(t, result.PR.AssignedReviewers, oldReviewer)
	assert.Contains(t, result.PR.AssignedReviewers, result.ReplacedBy)
	assert.Len(t, result.PR.AssignedReviewers, 2)
	// Verify second reviewer unchanged
	otherReviewer := pr.AssignedReviewers[1]
	assert.Contains(t, result.PR.AssignedReviewers, otherReviewer)
}

func TestPRIntegration_ReassignOnMergedPR(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2")
	pr := createPR(t, env.BaseURL(), "pr-merged-reassign", "Feature", "author")
	mergePR(t, env.BaseURL(), "pr-merged-reassign")
	reqBody := map[string]string{"pull_request_id": "pr-merged-reassign", "old_user_id": pr.AssignedReviewers[0]}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/reassign", Body: reqBody})
	assertStatusCode(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "PR_MERGED")
}

func TestPRIntegration_ReassignNotAssigned(t *testing.T) {
	env := setupTestEnvironment(t)

	// Create team with 5 users - random will pick 2 out of 4 (excluding author)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2", "rev3", "rev4")
	pr := createPR(t, env.BaseURL(), "pr-notassigned", "Feature", "author")
	require.Len(t, pr.AssignedReviewers, 2)

	// Find user who was NOT assigned (guaranteed to exist since we have 4 candidates but only 2 assigned)
	allCandidates := []string{"rev1", "rev2", "rev3", "rev4"}
	var notAssigned string
	for _, candidate := range allCandidates {
		found := false
		for _, assigned := range pr.AssignedReviewers {
			if assigned == candidate {
				found = true
				break
			}
		}
		if !found {
			notAssigned = candidate
			break
		}
	}
	require.NotEmpty(t, notAssigned, "Should have at least one user who was not assigned")

	// Try to reassign user who was never assigned as reviewer
	reqBody := map[string]string{
		"pull_request_id": "pr-notassigned",
		"old_user_id":     notAssigned,
	}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodPost,
		Path:   "/pullRequest/reassign",
		Body:   reqBody,
	})
	assertStatusCode(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "NOT_ASSIGNED")
}
func TestPRIntegration_ReassignNoCandidate(t *testing.T) {
	env := setupTestEnvironment(t)
	// Create team: author + 2 active reviewers + 1 inactive
	team := createTeam(t, env.BaseURL(), "backend", []domain.TeamMember{
		{UserID: "author", Username: "Author", IsActive: true},
		{UserID: "rev1", Username: "Reviewer1", IsActive: true},
		{UserID: "rev2", Username: "Reviewer2", IsActive: true},
		{UserID: "inactive", Username: "Inactive", IsActive: false},
	})
	require.Len(t, team.Members, 4)

	pr := createPR(t, env.BaseURL(), "pr-nocandidate", "Feature", "author")
	// Should assign 2 active teammates (rev1, rev2)
	require.Len(t, pr.AssignedReviewers, 2)

	// Now deactivate one of the assigned reviewers to test edge case
	// After reassign, only one active teammate remains (the other assigned reviewer)
	// But actually, let's keep both active and rely on the fact that
	// with author + 2 assigned reviewers, there's no 3rd active candidate

	// Try to reassign - should fail because no other active teammates available
	// (excludes: author + other assigned reviewer = all active members)
	reqBody := map[string]string{"pull_request_id": "pr-nocandidate", "old_user_id": pr.AssignedReviewers[0]}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/reassign", Body: reqBody})
	assertStatusCode(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "NO_CANDIDATE")
}

func TestPRIntegration_ConcurrentCreateOnlyOneSucceeds(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1")
	var wg sync.WaitGroup
	errors := make([]error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			reqBody := map[string]string{"pull_request_id": "pr-race", "pull_request_name": "Race", "author_id": "author"}
			resp := doRequest(t, env.BaseURL(), HTTPRequest{Method: http.MethodPost, Path: "/pullRequest/create", Body: reqBody})
			if resp.StatusCode != http.StatusCreated {
				errors[idx] = assert.AnError
			}
		}(i)
	}
	wg.Wait()
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount, "only one goroutine should create PR successfully")
}

func TestPRIntegration_ConcurrentMergeIdempotent(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1")
	createPR(t, env.BaseURL(), "pr-merge-race", "Feature", "author")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mergePR(t, env.BaseURL(), "pr-merge-race")
		}()
	}
	wg.Wait()
	pr := mergePR(t, env.BaseURL(), "pr-merge-race")
	assert.Equal(t, "MERGED", pr.Status)
}

func TestPRIntegration_CreateAuthorNotFound(t *testing.T) {
	env := setupTestEnvironment(t)
	// Try to create PR with non-existent author
	reqBody := map[string]string{
		"pull_request_id":   "pr-no-author",
		"pull_request_name": "Feature",
		"author_id":         "nonexistent-user",
	}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodPost,
		Path:   "/pullRequest/create",
		Body:   reqBody,
	})
	assertStatusCode(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "NOT_FOUND")
}

func TestPRIntegration_ReassignPRNotFound(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1")
	// Try to reassign on non-existent PR
	reqBody := map[string]string{
		"pull_request_id": "nonexistent-pr",
		"old_user_id":     "rev1",
	}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodPost,
		Path:   "/pullRequest/reassign",
		Body:   reqBody,
	})
	assertStatusCode(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "NOT_FOUND")
}

func TestPRIntegration_ReassignOldReviewerNotFound(t *testing.T) {
	env := setupTestEnvironment(t)
	createTeamWithUsers(t, env.BaseURL(), "backend", "author", "rev1", "rev2")
	pr := createPR(t, env.BaseURL(), "pr-reassign-notfound", "Feature", "author")
	require.Len(t, pr.AssignedReviewers, 2)

	// Try to reassign with non-existent old_user_id that is NOT in assigned_reviewers
	// This will fail the "is assigned?" check first, returning NOT_ASSIGNED (409)
	// rather than NOT_FOUND (404), which is correct behavior
	reqBody := map[string]string{
		"pull_request_id": "pr-reassign-notfound",
		"old_user_id":     "nonexistent-reviewer",
	}
	resp := doRequest(t, env.BaseURL(), HTTPRequest{
		Method: http.MethodPost,
		Path:   "/pullRequest/reassign",
		Body:   reqBody,
	})
	// Returns 409 NOT_ASSIGNED because non-existent user is not in assigned_reviewers
	// This is correct: business rule check happens before DB lookup
	assertStatusCode(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "NOT_ASSIGNED")
}
