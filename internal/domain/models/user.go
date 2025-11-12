package domain

import (
	dto "cw/internal/dto/models"
)

type User struct {
	ID                  string
	Name                string
	Email               string
	BirthDate           string
	PasswordHash        string
	IsAdmin             bool
	TwoFANeeded         bool
	Email2FACode        string
	Email2FAExpires     *string
	FailedLoginAttempts int
	LockedUntil         *string
}

type UserFilters struct {
	Name    string
	Email   string
	IsAdmin *bool
}

func UserFiltersFromDTO(dtoFilters dto.UserFilters) UserFilters {
	return UserFilters{
		Name:    dtoFilters.Name,
		Email:   dtoFilters.Email,
		IsAdmin: dtoFilters.IsAdmin,
	}
}

func CreateUserFromDTO(req dto.CreateUserRequest) User {
	return User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
		BirthDate:    req.BirthDate,
	}
}

func UpdateUserFromDTO(req dto.UpdateUserRequest) User {
	return User{
		Name:      req.Name,
		Email:     req.Email,
		BirthDate: req.BirthDate,
	}
}

func LoginFromDTO(req dto.LoginRequest) User {
	return User{
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}
}

func UserToDTO(domainUser User) dto.UserResponse {
	return dto.UserResponse{
		ID:        domainUser.ID,
		Name:      domainUser.Name,
		Email:     domainUser.Email,
		BirthDate: domainUser.BirthDate,
		IsAdmin:   domainUser.IsAdmin,
	}
}

func UsersToDTO(domainUsers []User) []dto.UserResponse {
	dtos := make([]dto.UserResponse, len(domainUsers))
	for i, user := range domainUsers {
		dtos[i] = UserToDTO(user)
	}
	return dtos
}

func AuthToDTO(id string, token string, message string, twoFANeeded bool) dto.AuthResponse {
	return dto.AuthResponse{
		UserID:      id,
		Token:       token,
		TwoFANeeded: twoFANeeded,
		Message:     message,
	}
}
