package service

import (
	"context"
	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll(ctx context.Context) ([]models.User, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (models.User, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetNicknameByID(ctx context.Context, id string) (string, *utils.Error) {
	return s.repo.GetNicknameByID(ctx, id)
}

func (s *UserService) GetAdminStatus(ctx context.Context) (models.User, *utils.Error) {
	return s.repo.GetAdminStatus(ctx)
}

func (s *UserService) UpdateAdminStatus(ctx context.Context, admin models.UserAdmin) (models.UserAdmin, *utils.Error) {
	return s.repo.UpdateAdminStatus(ctx, admin)
}

func (s *UserService) Update(ctx context.Context, user models.User) (models.User, *utils.Error) {
	return s.repo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) Login(ctx context.Context, user models.UserLogin) (models.User, *utils.Error) {
	return s.repo.Login(ctx, user)
}

func (s *UserService) Register(ctx context.Context, user models.UserRegister) (models.User, *utils.Error) {
	return s.repo.Register(ctx, user)
}
