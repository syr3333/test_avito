package dto

import (
	"avito/internal/domain"
	"strings"
)

type UserRequest struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
}

func (r *UserRequest) Validate() error {
	if err := ValidateUserID(r.ID); err != nil {
		return err
	}
	if err := ValidateUsername(r.Username); err != nil {
		return err
	}
	return nil
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func (r *SetIsActiveRequest) Validate() error {
	return ValidateUserID(r.UserID)
}

type UserResponse struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// ToDomain преобразует DTO в domain модель
func (r *UserRequest) ToDomain() *domain.User {
	return &domain.User{
		ID:       r.ID,
		Username: r.Username,
		IsActive: true,
	}
}

func UserFromDomain(user *domain.User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func ValidateUserID(id string) error {
	if strings.TrimSpace(id) == "" {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "user_id cannot be empty")
	}
	if len(id) > 255 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "user_id too long (max 255 characters)")
	}
	if !idRegex.MatchString(id) {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "user_id contains invalid characters")
	}
	return nil
}

func ValidateUsername(name string) error {
	if strings.TrimSpace(name) == "" {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "username cannot be empty")
	}
	if len(name) > 255 {
		return domain.NewAppError(domain.ErrCodeInvalidInput, "username too long (max 255 characters)")
	}
	return nil
}
