package domain

const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)

const (
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeInvalidRequest = "INVALID_REQUEST"
	ErrCodeInternalError  = "INTERNAL_ERROR"
	ErrCodeTeamExists     = "TEAM_EXISTS"
	ErrCodePRExists       = "PR_EXISTS"
	ErrCodePRMerged       = "PR_MERGED"
	ErrCodeNotAssigned    = "NOT_ASSIGNED"
	ErrCodeNoCandidate    = "NO_CANDIDATE"
	ErrCodeNotFound       = "NOT_FOUND"
)
