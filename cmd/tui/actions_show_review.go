package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) manageMovieShows() {
	for {
		fmt.Println("\n--- Movie Show Management ---")
		fmt.Println("1. List movie shows")
		fmt.Println("2. Get movie show by ID")
		fmt.Println("3. Create movie show")
		fmt.Println("4. Update movie show")
		fmt.Println("5. Delete movie show")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listMovieShows()
		case "2":
			app.getMovieShow()
		case "3":
			app.createMovieShow()
		case "4":
			app.updateMovieShow()
		case "5":
			app.deleteMovieShow()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listMovieShows() {
	page, limit := app.getPaginationParams()

	var filters dto.MovieShowFilters
	fmt.Print("Movie IDs (comma separated): ")
	app.scanner.Scan()
	movieIDs := app.scanner.Text()
	if movieIDs != "" {
		filters.MovieIDs = strings.Split(movieIDs, ",")
	}

	fmt.Print("Hall IDs (comma separated): ")
	app.scanner.Scan()
	hallIDs := app.scanner.Text()
	if hallIDs != "" {
		filters.HallIDs = strings.Split(hallIDs, ",")
	}

	fmt.Print("Date (YYYY-MM-DD): ")
	app.scanner.Scan()
	filters.Date = app.scanner.Text()

	result, err := app.movieShowService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d movie shows (total: %d):\n", len(result.Data), result.Total)
	for _, show := range result.Data {
		fmt.Printf("ID: %s, Movie: %s, Hall: %s, Start: %s\n",
			show.ID, show.MovieID, show.HallID, show.StartTime)
	}
}

func (app *TUIApp) getMovieShow() {
	fmt.Print("Enter movie show ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	show, err := app.movieShowService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", show.ID)
	fmt.Printf("Movie ID: %s\n", show.MovieID)
	fmt.Printf("Hall ID: %s\n", show.HallID)
	fmt.Printf("Start Time: %s\n", show.StartTime)
	fmt.Printf("Language: %s\n", show.Language)
}

func (app *TUIApp) createMovieShow() {
	var req dto.CreateMovieShowRequest

	fmt.Print("Movie ID: ")
	app.scanner.Scan()
	req.MovieID = app.scanner.Text()

	fmt.Print("Hall ID: ")
	app.scanner.Scan()
	req.HallID = app.scanner.Text()

	fmt.Print("Start Time (YYYY-MM-DD HH:MM:SS): ")
	app.scanner.Scan()
	startTimeStr := app.scanner.Text()

	startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		fmt.Printf("Invalid time format: %v\n", err)
		return
	}
	req.StartTime = startTime

	fmt.Print("Language: ")
	app.scanner.Scan()
	req.Language = app.scanner.Text()

	fmt.Print("Price: ")
	app.scanner.Scan()
	priceStr := app.scanner.Text()
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		req.BasePrice = price
	}

	id, err := app.movieShowService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Movie show created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateMovieShow() {
	fmt.Print("Enter movie show ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateMovieShowRequest
	fmt.Print("New movie ID: ")
	app.scanner.Scan()
	req.MovieID = app.scanner.Text()

	fmt.Print("New hall ID: ")
	app.scanner.Scan()
	req.HallID = app.scanner.Text()

	fmt.Print("New start time (YYYY-MM-DD HH:MM:SS): ")
	app.scanner.Scan()
	startTimeStr := app.scanner.Text()

	startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		fmt.Printf("Invalid time format: %v\n", err)
		return
	}
	req.StartTime = startTime

	fmt.Print("New language: ")
	app.scanner.Scan()
	req.Language = app.scanner.Text()

	fmt.Print("New price: ")
	app.scanner.Scan()
	priceStr := app.scanner.Text()
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		req.BasePrice = price
	}

	err = app.movieShowService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Movie show updated successfully!")
}
func (app *TUIApp) deleteMovieShow() {
	fmt.Print("Enter movie show ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.movieShowService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Movie show deleted successfully!")
}

func (app *TUIApp) manageReviews() {
	for {
		fmt.Println("\n--- Review Management ---")
		fmt.Println("1. List reviews")
		fmt.Println("2. Create review")
		fmt.Println("3. Update review")
		fmt.Println("4. Delete review")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listReviews()
		case "2":
			app.createReview()
		case "3":
			app.updateReview()
		case "4":
			app.deleteReview()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listReviews() {
	page, limit := app.getPaginationParams()

	var filters dto.ReviewFilters
	fmt.Print("Movie ID: ")
	app.scanner.Scan()
	filters.MovieID = app.scanner.Text()

	fmt.Print("User ID: ")
	app.scanner.Scan()
	filters.UserID = app.scanner.Text()

	result, err := app.reviewService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d reviews (total: %d):\n", len(result.Data), result.Total)
	for _, review := range result.Data {
		fmt.Printf("ID: %s, Movie: %s, User: %s, Rating: %d\n",
			review.ID, review.MovieID, review.UserID, review.Rating)
	}
}

func (app *TUIApp) createReview() {
	var req dto.CreateReviewRequest

	fmt.Print("Movie ID: ")
	app.scanner.Scan()
	movieID := app.scanner.Text()

	fmt.Print("Rating (1-10): ")
	app.scanner.Scan()
	ratingStr := app.scanner.Text()
	if rating, err := strconv.Atoi(ratingStr); err == nil {
		req.Rating = rating
	}

	fmt.Print("Comment: ")
	app.scanner.Scan()
	req.Comment = app.scanner.Text()

	id, err := app.reviewService.Create(context.Background(), req, movieID)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Review created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateReview() {
	fmt.Print("Enter review ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateReviewRequest
	fmt.Print("New rating (1-10): ")
	app.scanner.Scan()
	ratingStr := app.scanner.Text()
	if rating, err := strconv.Atoi(ratingStr); err == nil {
		req.Rating = rating
	}

	fmt.Print("New comment: ")
	app.scanner.Scan()
	req.Comment = app.scanner.Text()

	err := app.reviewService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Review updated successfully!")
}

func (app *TUIApp) deleteReview() {
	fmt.Print("Enter review ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.reviewService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Review deleted successfully!")
}
