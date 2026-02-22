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

type reviewService struct {
	client *client.APIClient
}

func NewReviewService(apiClient *client.APIClient) ReviewService {
	return &reviewService{client: apiClient}
}

func (s *reviewService) GetAll(ctx context.Context, filters dto.ReviewFilters, page, limit int) (*dto.PaginatedReviewResponse, error) {
	path := s.buildGetAllPath(filters, page, limit)

	resp, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Ошибка при закрытии тела ответа: %v", err)
		}
	}()

	return s.handleGetAllResponse(resp)
}

// buildGetAllPath constructs the path with query parameters
func (s *reviewService) buildGetAllPath(filters dto.ReviewFilters, page, limit int) string {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("limit", strconv.Itoa(limit))

	if filters.MovieID != "" {
		params.Add("movie_id", filters.MovieID)
	}
	if filters.UserID != "" {
		params.Add("user_id", filters.UserID)
	}
	if filters.Comment != "" {
		params.Add("comment", filters.Comment)
	}
	if filters.RatingMin > 0 {
		params.Add("rating_min", strconv.Itoa(filters.RatingMin))
	}
	if filters.RatingMax > 0 {
		params.Add("rating_max", strconv.Itoa(filters.RatingMax))
	}

	path := "/reviews"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	return path
}

// handleGetAllResponse processes the HTTP response
func (s *reviewService) handleGetAllResponse(resp *http.Response) (*dto.PaginatedReviewResponse, error) {
	if resp.StatusCode != http.StatusOK {
		return s.handleErrorStatusCode(resp)
	}

	var result dto.PaginatedReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// handleErrorStatusCode handles non-200 status codes
func (s *reviewService) handleErrorStatusCode(resp *http.Response) (*dto.PaginatedReviewResponse, error) {
	var errorResp dto.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
		return nil, fmt.Errorf("API error: %s", errorResp.Message)
	}
	return nil, fmt.Errorf("API error: %s", resp.Status)
}

func (s *reviewService) Create(ctx context.Context, review dto.CreateReviewRequest, movieID string) (string, error) {
	path := fmt.Sprintf("/movies/%s/reviews", movieID)

	resp, err := s.client.DoRequest("POST", path, review)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Ошибка при закрытии тела ответа: %v", err)
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

func (s *reviewService) Update(ctx context.Context, id string, review dto.UpdateReviewRequest) error {
	path := fmt.Sprintf("/reviews/%s", id)

	resp, err := s.client.DoRequest("PUT", path, review)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Ошибка при закрытии тела ответа: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("review not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

func (s *reviewService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/reviews/%s", id)

	resp, err := s.client.DoRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Ошибка при закрытии тела ответа: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("review not found")
		}
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Message)
		}
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}
