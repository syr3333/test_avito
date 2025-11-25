package dto

import (
	"avito/internal/domain"
	"testing"
)

func TestValidateTeamName(t *testing.T) {
	tests := []struct {
		name      string
		teamName  string
		wantError bool
	}{
		{"valid name", "backend-team", false},
		{"valid with spaces", "Backend Team", false},
		{"empty string", "", true},
		{"too long", string(make([]byte, 256)), true},
		{"invalid chars", "team@name!", true},
		{"valid with underscore", "team_name_1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamName(tt.teamName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTeamName(%q) error = %v, wantError %v", tt.teamName, err, tt.wantError)
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		wantError bool
	}{
		{"valid id", "user123", false},
		{"valid with dash", "user-123", false},
		{"valid with underscore", "user_123", false},
		{"empty string", "", true},
		{"too long", string(make([]byte, 256)), true},
		{"invalid chars", "user@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserID(tt.userID)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateUserID(%q) error = %v, wantError %v", tt.userID, err, tt.wantError)
			}
		})
	}
}

func TestValidatePullRequestID(t *testing.T) {
	tests := []struct {
		name      string
		prID      string
		wantError bool
	}{
		{"valid id", "pr-1001", false},
		{"valid simple", "pr1", false},
		{"empty string", "", true},
		{"too long", string(make([]byte, 256)), true},
		{"invalid chars", "pr@1001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePullRequestID(tt.prID)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePullRequestID(%q) error = %v, wantError %v", tt.prID, err, tt.wantError)
			}
		})
	}
}

func TestValidateTeamRequest(t *testing.T) {
	tests := []struct {
		name      string
		team      *TeamRequest
		wantError bool
	}{
		{
			name: "valid team",
			team: &TeamRequest{
				Name: "backend",
				Members: []TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				},
			},
			wantError: false,
		},
		{
			name: "empty members",
			team: &TeamRequest{
				Name:    "backend",
				Members: []TeamMember{},
			},
			wantError: true,
		},
		{
			name: "invalid team name",
			team: &TeamRequest{
				Name: "",
				Members: []TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				},
			},
			wantError: true,
		},
		{
			name: "invalid member id",
			team: &TeamRequest{
				Name: "backend",
				Members: []TeamMember{
					{UserID: "", Username: "Alice", IsActive: true},
				},
			},
			wantError: true,
		},
		{
			name: "too many members",
			team: &TeamRequest{
				Name:    "backend",
				Members: make([]TeamMember, 201),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamRequest(tt.team)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTeamRequest() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewAppError(t *testing.T) {
	err := domain.NewAppError(domain.ErrCodeNotFound, "test message")

	if err.Code != domain.ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", domain.ErrCodeNotFound, err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}

	if err.Error() != "test message" {
		t.Errorf("Expected Error() to return 'test message', got '%s'", err.Error())
	}
}
