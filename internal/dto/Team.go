package dto

import (
	"avito/internal/domain"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Regex для валидации ID (буквы, цифры, дефис, подчеркивание)
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	// Regex для имени команды/пользователя
	nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\s-]+$`)
)

// TeamRequest - входящий запрос для создания команды
type TeamRequest struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

// Validate проверяет корректность данных в запросе
func (r *TeamRequest) Validate() error {
	return ValidateTeamRequest(r)
}

// TeamResponse - ответ с данными команды
type TeamResponse struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// MassDeactivateRequest - запрос на массовую деактивацию
type MassDeactivateRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}

// Validate проверяет корректность запроса
func (r *MassDeactivateRequest) Validate() error {
	if err := ValidateTeamName(r.TeamName); err != nil {
		return err
	}
	if len(r.UserIDs) == 0 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "user_ids cannot be empty")
	}
	for i, uid := range r.UserIDs {
		if err := ValidateUserID(uid); err != nil {
			return domain.NewAppError(domain.ErrCodeInvalidInput, fmt.Sprintf("user_ids[%d]: %s", i, err.Error()))
		}
	}
	return nil
}

// ToDomain преобразует DTO в domain модель
func (r *TeamRequest) ToDomain() *domain.Team {
	members := make([]domain.TeamMember, len(r.Members))
	for i, m := range r.Members {
		members[i] = domain.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}
	return &domain.Team{
		Name:    r.Name,
		Members: members,
	}
}

// FromDomain преобразует domain модель в DTO
func TeamFromDomain(team *domain.Team) TeamResponse {
	members := make([]TeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}
	return TeamResponse{
		TeamName: team.Name,
		Members:  members,
	}
}

func ValidateTeamRequest(req *TeamRequest) error {
	if err := ValidateTeamName(req.Name); err != nil {
		return err
	}

	if len(req.Members) == 0 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "team must have at least one member")
	}

	if len(req.Members) > 200 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, fmt.Sprintf("team has too many members (max 200, got %d)", len(req.Members)))
	}

	for i, member := range req.Members {
		if err := ValidateUserID(member.UserID); err != nil {
			return domain.NewAppError(domain.ErrCodeInvalidInput, fmt.Sprintf("member[%d]: %s", i, err.Error()))
		}
		if err := ValidateUsername(member.Username); err != nil {
			return domain.NewAppError(domain.ErrCodeInvalidInput, fmt.Sprintf("member[%d]: %s", i, err.Error()))
		}
	}

	return nil
}

func ValidateTeamName(name string) error {
	if strings.TrimSpace(name) == "" {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "team_name cannot be empty")
	}
	if len(name) > 255 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "team_name too long (max 255 characters)")
	}
	if !nameRegex.MatchString(name) {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "team_name contains invalid characters")
	}
	return nil
}
