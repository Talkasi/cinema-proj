package services

import (
	"context"
	dto "cw/internal/dto/models"
	"cw/internal/tui/client"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type ticketService struct {
	client *client.APIClient
}

func NewTicketService(apiClient *client.APIClient) TicketService {
	return &ticketService{client: apiClient}
}

func (s *ticketService) GetAll(ctx context.Context, filters dto.TicketFilters, page, limit int) (*dto.PaginatedTicketResponse, error) {
	path := s.buildGetAllPath(filters, page, limit)

	resp, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	return s.handleGetAllResponse(resp)
}

// buildGetAllPath constructs the path with query parameters
func (s *ticketService) buildGetAllPath(filters dto.TicketFilters, page, limit int) string {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("limit", strconv.Itoa(limit))

	for _, Status := range filters.Status {
		params.Add("ticket_Status", Status)
	}
	if filters.MovieShowID != "" {
		params.Add("movie_show_id", filters.MovieShowID)
	}
	for _, seatID := range filters.SeatID {
		params.Add("seat_id", seatID)
	}
	for _, userID := range filters.UserID {
		params.Add("user_id", userID)
	}
	if filters.PriceMin > 0 {
		params.Add("price_min", fmt.Sprintf("%.2f", filters.PriceMin))
	}
	if filters.PriceMax > 0 {
		params.Add("price_max", fmt.Sprintf("%.2f", filters.PriceMax))
	}

	path := "/tickets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	return path
}

// handleGetAllResponse processes the HTTP response
func (s *ticketService) handleGetAllResponse(resp *http.Response) (*dto.PaginatedTicketResponse, error) {
	if resp.StatusCode != http.StatusOK {
		return s.handleErrorStatusCode(resp)
	}

	var result dto.PaginatedTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// handleErrorStatusCode handles non-200 status codes
func (s *ticketService) handleErrorStatusCode(resp *http.Response) (*dto.PaginatedTicketResponse, error) {
	var errorResp dto.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
		return nil, fmt.Errorf("API error: %s", errorResp.Message)
	}
	return nil, fmt.Errorf("API error: %s", resp.Status)
}

func (s *ticketService) GetByID(ctx context.Context, id string) (*dto.TicketResponse, error) {
	path := fmt.Sprintf("/tickets/%s", id)

	resp, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *ticketService) Create(ctx context.Context, ticket dto.CreateTicketRequest, movieShowID string) (string, error) {
	path := fmt.Sprintf("/movie-shows/%s/tickets", movieShowID)

	resp, err := s.client.DoRequest("POST", path, ticket)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return "", fmt.Errorf("API error: %s", errorResp.Message)
		}
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.CreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

func (s *ticketService) UpdateStatus(ctx context.Context, id string, status dto.UpdateStatusRequest) error {
	path := fmt.Sprintf("/tickets/%s", id)

	resp, err := s.client.DoRequest("PATCH", path, status)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("ticket not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

func (s *ticketService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/tickets/%s", id)

	resp, err := s.client.DoRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("ticket not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}
