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

type movieShowService struct {
	client *client.APIClient
}

func NewMovieShowService(apiClient *client.APIClient) MovieShowService {
	return &movieShowService{client: apiClient}
}

func (s *movieShowService) GetAll(ctx context.Context, filters dto.MovieShowFilters, page, limit int) (*dto.PaginatedMovieShowResponse, error) {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("limit", strconv.Itoa(limit))

	for _, movieID := range filters.MovieIDs {
		params.Add("movie_id", movieID)
	}
	for _, hallID := range filters.HallIDs {
		params.Add("hall_id", hallID)
	}
	for _, language := range filters.Languages {
		params.Add("language", language)
	}
	if filters.Date != "" {
		params.Add("date", filters.Date)
	}
	if filters.StartTimeFrom != "" {
		params.Add("start_time_from", filters.StartTimeFrom)
	}
	if filters.StartTimeTo != "" {
		params.Add("start_time_to", filters.StartTimeTo)
	}

	path := "/movie-shows"
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

	var result dto.PaginatedMovieShowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *movieShowService) GetByID(ctx context.Context, id string) (*dto.MovieShowResponse, error) {
	path := fmt.Sprintf("/movie-shows/%s", id)

	resp, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("movie show not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.MovieShowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *movieShowService) Create(ctx context.Context, show dto.CreateMovieShowRequest) (string, error) {
	resp, err := s.client.DoRequest("POST", "/movie-shows", show)
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

func (s *movieShowService) Update(ctx context.Context, id string, show dto.UpdateMovieShowRequest) error {
	path := fmt.Sprintf("/movie-shows/%s", id)

	resp, err := s.client.DoRequest("PUT", path, show)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("movie show not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

func (s *movieShowService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/movie-shows/%s", id)

	resp, err := s.client.DoRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("movie show not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}
