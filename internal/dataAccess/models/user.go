package entity

import (
	domain "cw/internal/domain/models"
	"time"
)

type User struct {
	ID                  string     `db:"id"`
	Name                string     `db:"name"`
	Email               string     `db:"email"`
	PasswordHash        string     `db:"password_hash"`
	BirthDate           time.Time  `db:"birth_date"`
	IsAdmin             bool       `db:"is_admin"`
	TwoFANeeded         bool       `db:"two_fa_enabled"`
	Email2FACode        *string    `db:"email_2fa_code"`
	Email2FAExpires     *time.Time `db:"email_2fa_expires"`
	FailedLoginAttempts int        `db:"failed_login_attempts"`
	LockedUntil         *time.Time `db:"locked_until"`
}

func UserToDomain(entity User) domain.User {
	var lockedUntilStr *string
	if entity.LockedUntil != nil {
		dateStr := entity.LockedUntil.Format("2006-01-02T15:04:05Z07:00")
		lockedUntilStr = &dateStr
	}

	var email2FAExpiresStr *string
	if entity.Email2FAExpires != nil {
		dateStr := entity.Email2FAExpires.Format("2006-01-02T15:04:05Z07:00")
		email2FAExpiresStr = &dateStr
	}

	var email2FACode string
	if entity.Email2FACode != nil {
		email2FACode = *entity.Email2FACode
	}

	return domain.User{
		ID:           entity.ID,
		Name:         entity.Name,
		Email:        entity.Email,
		BirthDate:    entity.BirthDate.Format("2006-01-02"),
		PasswordHash: entity.PasswordHash,
		IsAdmin:     entity.IsAdmin,
		TwoFANeeded: entity.TwoFANeeded,

		Email2FACode:        email2FACode,
		Email2FAExpires:     email2FAExpiresStr,
		FailedLoginAttempts: entity.FailedLoginAttempts,
		LockedUntil:         lockedUntilStr,
	}
}

func UserFromDomain(domainUser domain.User) User {
	birthDate, _ := time.Parse("2006-01-02", domainUser.BirthDate)

	var lockedUntil *time.Time
	if domainUser.LockedUntil != nil {
		parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", *domainUser.LockedUntil)
		if err == nil {
			lockedUntil = &parsedTime
		}
	}

	var email2FAExpires *time.Time
	if domainUser.Email2FAExpires != nil {
		parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", *domainUser.Email2FAExpires)
		if err == nil {
			email2FAExpires = &parsedTime
		}
	}

	var email2FACode *string
	if domainUser.Email2FACode != "" {
		email2FACode = &domainUser.Email2FACode
	}

	return User{
		ID:           domainUser.ID,
		Name:         domainUser.Name,
		Email:        domainUser.Email,
		BirthDate:    birthDate,
		PasswordHash: domainUser.PasswordHash,
		IsAdmin:     domainUser.IsAdmin,
		TwoFANeeded: domainUser.TwoFANeeded,

		Email2FACode:        email2FACode,
		Email2FAExpires:     email2FAExpires,
		FailedLoginAttempts: domainUser.FailedLoginAttempts,
		LockedUntil:         lockedUntil,
	}
}

func UsersToDomain(entities []User) []domain.User {
	domainUsers := make([]domain.User, len(entities))
	for i, entity := range entities {
		domainUsers[i] = UserToDomain(entity)
	}
	return domainUsers
}
