package service

import (
	"context"

	"cw/internal/domain"
	"cw/internal/repository"
)

type TicketService struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) GetByMovieShowID(ctx context.Context, movieShowId string) ([]domain.Ticket, error) {
	return s.repo.GetByMovieShowID(ctx, movieShowId)
}

func (s *TicketService) GetAvailableByMovieShowID(ctx context.Context, movieShowId string) ([]domain.Ticket, error) {
	return s.repo.GetAvailableByMovieShowID(ctx, movieShowId)
}

func (s *TicketService) GetByUserId(ctx context.Context, userId string) ([]domain.Ticket, error) {
	return s.repo.GetByUserId(ctx, userId)
}

func (s *TicketService) GetByID(ctx context.Context, id string) (domain.Ticket, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TicketService) Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	return s.repo.Create(ctx, ticket)
}

func (s *TicketService) Update(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	return s.repo.Update(ctx, ticket)
}

func (s *TicketService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
