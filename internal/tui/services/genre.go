package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	dto "cw/internal/dto/models"
	"cw/internal/tui/client"
)

type genreService struct {
	client *client.APIClient
}

func NewGenreService(apiClient *client.APIClient) GenreService {
	return &genreService{client: apiClient}
}

func (s *genreService) GetAll(ctx context.Context, filters dto.GenreFilters, page, limit int) (*dto.PaginatedGenreResponse, error) {

	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("limit", strconv.Itoa(limit))

	if filters.Name != "" {
		params.Add("name", filters.Name)
	}
	if filters.Description != "" {
		params.Add("description", filters.Description)
	}

	path := "/genres"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

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
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.PaginatedGenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *genreService) GetByID(ctx context.Context, id string) (*dto.GenreResponse, error) {
	path := fmt.Sprintf("/genres/%s", id)

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
			return nil, fmt.Errorf("genre not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result dto.GenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *genreService) Create(ctx context.Context, genre dto.CreateGenreRequest) (string, error) {
	resp, err := s.client.DoRequest("POST", "/genres", genre)
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

func (s *genreService) Update(ctx context.Context, id string, genre dto.UpdateGenreRequest) error {
	path := fmt.Sprintf("/genres/%s", id)

	resp, err := s.client.DoRequest("PUT", path, genre)
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
			return fmt.Errorf("genre not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

func (s *genreService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/genres/%s", id)

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
			return fmt.Errorf("genre not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}
