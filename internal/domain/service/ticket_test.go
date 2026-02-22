package service

import (
	"context"
	"errors"
	"testing"

	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type ticketRepoMock struct {
	getAllFn             func(context.Context, domain.TicketFilters, int, int) ([]domain.Ticket, int, *utils.Error)
	getByIDFn            func(context.Context, string) (domain.Ticket, *utils.Error)
	createForMovieShowFn func(context.Context, string, domain.Ticket) (domain.Ticket, *utils.Error)
	updateStatusFn       func(context.Context, string, domain.Ticket) (domain.Ticket, *utils.Error)
	deleteFn             func(context.Context, string) *utils.Error
}

func (m ticketRepoMock) GetAll(ctx context.Context, f domain.TicketFilters, page, limit int) ([]domain.Ticket, int, *utils.Error) {
	return m.getAllFn(ctx, f, page, limit)
}
func (m ticketRepoMock) GetByID(ctx context.Context, id string) (domain.Ticket, *utils.Error) {
	return m.getByIDFn(ctx, id)
}
func (m ticketRepoMock) CreateForMovieShow(ctx context.Context, movieShowID string, t domain.Ticket) (domain.Ticket, *utils.Error) {
	return m.createForMovieShowFn(ctx, movieShowID, t)
}
func (m ticketRepoMock) UpdateStatus(ctx context.Context, id string, status domain.Ticket) (domain.Ticket, *utils.Error) {
	return m.updateStatusFn(ctx, id, status)
}
func (m ticketRepoMock) Delete(ctx context.Context, id string) *utils.Error {
	return m.deleteFn(ctx, id)
}

func TestTicketServiceCreateForMovieShowPropagatesConflictAndPassesArgs(t *testing.T) {
	t.Parallel()

	expected := utils.NewConflict("Database conflict", errors.New("seat already booked"))
	inTicket := domain.Ticket{SeatID: "s1", Price: 500, Status: "reserved"}
	repo := ticketRepoMock{
		getAllFn: func(context.Context, domain.TicketFilters, int, int) ([]domain.Ticket, int, *utils.Error) {
			return nil, 0, nil
		},
		getByIDFn: func(context.Context, string) (domain.Ticket, *utils.Error) { return domain.Ticket{}, nil },
		createForMovieShowFn: func(_ context.Context, movieShowID string, ticket domain.Ticket) (domain.Ticket, *utils.Error) {
			if movieShowID != "show-1" {
				t.Fatalf("movieShowID = %q, want show-1", movieShowID)
			}
			if ticket.SeatID != "s1" || ticket.Status != "reserved" {
				t.Fatalf("unexpected ticket payload: %+v", ticket)
			}
			return domain.Ticket{}, expected
		},
		updateStatusFn: func(context.Context, string, domain.Ticket) (domain.Ticket, *utils.Error) {
			return domain.Ticket{}, nil
		},
		deleteFn: func(context.Context, string) *utils.Error { return nil },
	}

	svc := NewTicketService(repo)
	_, err := svc.CreateForMovieShow(context.Background(), "show-1", inTicket)
	if err != expected {
		t.Fatalf("error pointer was not propagated: got %p want %p", err, expected)
	}
}

func TestTicketServiceGetAllBuildsPaginatedResponse(t *testing.T) {
	t.Parallel()

	repo := ticketRepoMock{
		getAllFn: func(_ context.Context, _ domain.TicketFilters, page, limit int) ([]domain.Ticket, int, *utils.Error) {
			if page != 3 || limit != 25 {
				t.Fatalf("page/limit = %d/%d, want 3/25", page, limit)
			}
			return []domain.Ticket{{ID: "t1"}}, 72, nil
		},
		getByIDFn: func(context.Context, string) (domain.Ticket, *utils.Error) { return domain.Ticket{}, nil },
		createForMovieShowFn: func(context.Context, string, domain.Ticket) (domain.Ticket, *utils.Error) {
			return domain.Ticket{}, nil
		},
		updateStatusFn: func(context.Context, string, domain.Ticket) (domain.Ticket, *utils.Error) {
			return domain.Ticket{}, nil
		},
		deleteFn: func(context.Context, string) *utils.Error { return nil },
	}

	svc := NewTicketService(repo)
	resp, err := svc.GetAll(context.Background(), domain.TicketFilters{}, 3, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 72 || resp.Page != 3 || resp.Limit != 25 {
		t.Fatalf("unexpected pagination response: %+v", resp)
	}
}
