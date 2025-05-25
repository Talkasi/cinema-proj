package domain

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	BirthDate    string
	IsAdmin      bool
}
