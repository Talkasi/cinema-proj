package models

type User struct {
	ID           string `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Name         string `json:"name" example:"Иван Иванов"`
	Email        string `json:"email" example:"ivan@example.com"`
	PasswordHash string `json:"password_hash" example:"93652657623450"`
	BirthDate    string `json:"birth_date" example:"1990-01-01"`
	IsAdmin      bool   `json:"is_admin,omitempty" example:"true"`
}

type UserData struct {
	Name         string `json:"name" example:"Иван Иванов"`
	Email        string `json:"email" example:"ivan@example.com"`
	PasswordHash string `json:"password_hash" example:"93652657623450"`
	BirthDate    string `json:"birth_date" example:"1990-01-01"`
}

type UserLogin struct {
	Email        string `json:"email" example:"admin@admin.com"`
	PasswordHash string `json:"password_hash" example:"$2a$10$xS.xH8z3bJ1J5hNtGvXZfez7v6JQY9W7kZf3JvYbW6cXrV1nYd2E3C"`
}

type UserAdmin struct {
	IsAdmin bool `json:"is_admin" example:"true"`
}

type UserRegister struct {
	Name         string `json:"name" example:"Иван Иванов"`
	Email        string `json:"email" example:"ivan@example.com"`
	PasswordHash string `json:"password_hash" example:"hashed_password"`
	BirthDate    string `json:"birth_date" example:"1990-01-01"`
}
