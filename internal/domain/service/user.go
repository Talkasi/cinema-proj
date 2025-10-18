package service

import (
	"context"
	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Login(ctx context.Context, credentials domain.User) (domain.AuthResponse, *utils.Error) {
	return s.repo.Login(ctx, credentials)
}

func (s *UserService) Register(ctx context.Context, user domain.User) (domain.User, *utils.Error) {
	return s.repo.Register(ctx, user)
}

func (s *UserService) GetAll(ctx context.Context, filters domain.UserFilters, page, limit int) (domain.PaginatedResponse[domain.User], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.User]{}, err
	}

	return domain.PaginatedResponse[domain.User]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (domain.User, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id string, userData domain.User) (domain.User, *utils.Error) {
	return s.repo.Update(ctx, id, userData)
}

func (s *UserService) UpdateAdminStatus(ctx context.Context, id string, isAdmin bool) (domain.User, *utils.Error) {
	return s.repo.UpdateAdminStatus(ctx, id, isAdmin)
}

func (s *UserService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
