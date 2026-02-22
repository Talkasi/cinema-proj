package entity

import (
	domain "cw/internal/domain/models"
	"time"
)

type User struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	BirthDate    time.Time `db:"birth_date"`
	IsAdmin      bool      `db:"is_admin"`
}

func UserToDomain(entity User) domain.User {
	return domain.User{
		ID:           entity.ID,
		Name:         entity.Name,
		Email:        entity.Email,
		BirthDate:    entity.BirthDate.Format("2006-01-02"),
		PasswordHash: entity.PasswordHash,
		IsAdmin:      entity.IsAdmin,
	}
}

func UserFromDomain(domainUser domain.User) User {
	birthDate, _ := time.Parse("2006-01-02", domainUser.BirthDate)

	return User{
		ID:           domainUser.ID,
		Name:         domainUser.Name,
		Email:        domainUser.Email,
		BirthDate:    birthDate,
		PasswordHash: domainUser.PasswordHash,
		IsAdmin:      domainUser.IsAdmin,
	}
}

func UsersToDomain(entities []User) []domain.User {
	domainUsers := make([]domain.User, len(entities))
	for i, entity := range entities {
		domainUsers[i] = UserToDomain(entity)
	}
	return domainUsers
}
