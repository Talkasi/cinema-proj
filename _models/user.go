package models

import "time"

type UserRegistration struct {
	Name         string    `json:"name" validate:"required,min=2,max=50,nameFormat"`
	Email        string    `json:"email" validate:"required,email,max=100"`
	PasswordHash string    `json:"password_hash" validate:"required,min=8,max=72"`
	BirthDate    time.Time `json:"birth_date" validate:"required,validBirthDate"`
}

type UserLogin struct {
	Email        string `json:"email" validate:"required,email"`
	PasswordHash string `json:"password_hash" validate:"required"`
}

type UserProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	BirthDate time.Time `json:"birth_date"`
	IsBlocked bool      `json:"is_blocked"`
	IsAdmin   bool      `json:"is_admin"`
}

type UserUpdate struct {
	Name      *string    `json:"name,omitempty" validate:"omitempty,min=2,max=50,nameFormat"`
	Email     *string    `json:"email,omitempty" validate:"omitempty,email,max=100"`
	BirthDate *time.Time `json:"birth_date,omitempty" validate:"omitempty,validBirthDate"`
}

type UserAdminUpdate struct {
	IsAdmin *bool `json:"is_admin,omitempty"`
}

type UserBlockUpdate struct {
	IsBlocked *bool `json:"is_blocked,omitempty"`
}
