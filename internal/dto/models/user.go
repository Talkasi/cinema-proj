package dto

type UserFilters struct {
	Name    string `form:"name" example:"ivan"`
	Email   string `form:"email" example:"user@example.com"`
	IsAdmin *bool  `form:"is_admin" example:"true"`
}

type RegisterResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UserResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string `json:"name" example:"Ivan Ivanov"`
	Email     string `json:"email" example:"ivan@example.com"`
	BirthDate string `json:"birth_date" example:"1990-05-15"`
	IsAdmin   bool   `json:"is_admin" example:"false"`
}

type CreateUserRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=50" example:"Mariya Petrova"`
	Email        string `json:"email" validate:"required,email" example:"maria@example.com"`
	PasswordHash string `json:"password_hash" validate:"required,min=6" example:"securepassword123"`
	BirthDate    string `json:"birth_date" validate:"required" example:"1985-08-20"`
}

type UpdateUserRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=50" example:"Mariya Sidorova"`
	Email     string `json:"email" validate:"required,email" example:"maria.sidorova@example.com"`
	BirthDate string `json:"birth_date" validate:"required" example:"1985-08-20"`
}

type LoginRequest struct {
	Email        string `json:"email" validate:"required,email" example:"maria@example.com"`
	PasswordHash string `json:"password_hash" validate:"required" example:"securepassword123"`
}

type UpdateAdminStatusRequest struct {
	IsAdmin bool `json:"is_admin" validate:"required" example:"true"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" validate:"required,min=6" example:"newpassword123"`
}
