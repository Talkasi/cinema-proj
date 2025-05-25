package service

import (
	"context"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type TicketService struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) GetByMovieShowID(ctx context.Context, movieShowId string) ([]models.Ticket, *utils.Error) {
	return s.repo.GetByMovieShowID(ctx, movieShowId)
}

func (s *TicketService) GetAvailableByMovieShowID(ctx context.Context, movieShowId string) ([]models.Ticket, *utils.Error) {
	return s.repo.GetAvailableByMovieShowID(ctx, movieShowId)
}

func (s *TicketService) GetByUserId(ctx context.Context, userId string) ([]models.Ticket, *utils.Error) {
	return s.repo.GetByUserId(ctx, userId)
}

func (s *TicketService) GetByID(ctx context.Context, id string) (models.Ticket, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TicketService) Create(ctx context.Context, ticket models.Ticket) (models.Ticket, *utils.Error) {
	return s.repo.Create(ctx, ticket)
}

func (s *TicketService) Update(ctx context.Context, ticket models.Ticket) (models.Ticket, *utils.Error) {
	return s.repo.Update(ctx, ticket)
}

func (s *TicketService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
