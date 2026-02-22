package services

import (
	"context"
	dto "cw/internal/dto/models"
	"cw/internal/tui/client"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type seatService struct {
	client *client.APIClient
}

func NewSeatService(apiClient *client.APIClient) SeatService {
	return &seatService{client: apiClient}
}

func (s *seatService) GetByHall(ctx context.Context, hallID string, filters dto.SeatFilters, page, limit int) (*dto.PaginatedSeatResponse, error) {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("limit", strconv.Itoa(limit))

	if filters.SeatTypeID != "" {
		params.Add("seat_type_id", filters.SeatTypeID)
	}
	if filters.RowNumberMin > 0 {
		params.Add("row_number_min", strconv.Itoa(filters.RowNumberMin))
	}
	if filters.RowNumberMax > 0 {
		params.Add("row_number_max", strconv.Itoa(filters.RowNumberMax))
	}
	if filters.SeatNumberMin > 0 {
		params.Add("seat_number_min", strconv.Itoa(filters.SeatNumberMin))
	}
	if filters.SeatNumberMax > 0 {
		params.Add("seat_number_max", strconv.Itoa(filters.SeatNumberMax))
	}

	path := fmt.Sprintf("/halls/%s/seats", hallID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.PaginatedSeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *seatService) GetByID(ctx context.Context, id string) (*dto.SeatResponse, error) {
	path := fmt.Sprintf("/seats/%s", id)

	resp, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("seat not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.SeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *seatService) Create(ctx context.Context, seat dto.CreateSeatRequest) (string, error) {
	path := fmt.Sprintf("/halls/%s/seats", seat.HallID)

	resp, err := s.client.DoRequest("POST", path, seat)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

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

func (s *seatService) Update(ctx context.Context, id string, seat dto.UpdateSeatRequest) error {
	path := fmt.Sprintf("/seats/%s", id)

	resp, err := s.client.DoRequest("PUT", path, seat)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("seat not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

func (s *seatService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/seats/%s", id)

	resp, err := s.client.DoRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("seat not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}
