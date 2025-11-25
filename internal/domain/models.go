package domain

import "time"

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   string
}

type Team struct {
	Name    string
	Members []TeamMember
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}

type AssignmentStat struct {
	ID    string
	Count int
}

type Statistics struct {
	AssignmentsByUser []AssignmentStat
	AssignmentsByPR   []AssignmentStat
	TotalPRs          int
	TotalAssignments  int
	ActiveUsers       int
	Teams             int
}

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
