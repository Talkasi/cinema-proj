package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type TicketService struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) GetAll(ctx context.Context, filters domain.TicketFilters, page, limit int) (domain.PaginatedResponse[domain.Ticket], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.Ticket]{}, err
	}

	return domain.PaginatedResponse[domain.Ticket]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *TicketService) GetByID(ctx context.Context, id string) (domain.Ticket, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TicketService) CreateForMovieShow(ctx context.Context, movieShowId string, ticket domain.Ticket) (domain.Ticket, *utils.Error) {
	return s.repo.CreateForMovieShow(ctx, movieShowId, ticket)
}

func (s *TicketService) UpdateStatus(ctx context.Context, id string, StatusData domain.Ticket) (domain.Ticket, *utils.Error) {
	return s.repo.UpdateStatus(ctx, id, StatusData)
}

func (s *TicketService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
