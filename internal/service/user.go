package service

import (
	"context"
	"cw/internal/domain"
	"cw/internal/repository"
	"cw/internal/utils"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll(ctx context.Context) ([]domain.User, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (domain.User, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetNicknameByID(ctx context.Context, id string) (string, *utils.Error) {
	return s.repo.GetNicknameByID(ctx, id)
}

func (s *UserService) GetAdminStatus(ctx context.Context) (domain.User, *utils.Error) {
	return s.repo.GetAdminStatus(ctx)
}

func (s *UserService) UpdateAdminStatus(ctx context.Context, admin domain.User) (domain.User, *utils.Error) {
	return s.repo.UpdateAdminStatus(ctx, admin)
}

func (s *UserService) Update(ctx context.Context, id string, user domain.User) (domain.User, *utils.Error) {
	return s.repo.Update(ctx, id, user)
}

func (s *UserService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) Login(ctx context.Context, user domain.User) (domain.User, *utils.Error) {
	return s.repo.Login(ctx, user)
}

func (s *UserService) Register(ctx context.Context, user domain.User) (domain.User, *utils.Error) {
	return s.repo.Register(ctx, user)
}
