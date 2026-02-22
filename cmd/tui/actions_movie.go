package main

import (
	"context"
	"fmt"
	"strings"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) manageMovies() {
	for {
		fmt.Println("\n--- Movie Management ---")
		fmt.Println("1. List movies")
		fmt.Println("2. Get movie by ID")
		fmt.Println("3. Create movie")
		fmt.Println("4. Update movie")
		fmt.Println("5. Delete movie")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listMovies()
		case "2":
			app.getMovie()
		case "3":
			app.createMovie()
		case "4":
			app.updateMovie()
		case "5":
			app.deleteMovie()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listMovies() {
	page, limit := app.getPaginationParams()

	var filters dto.MovieFilters
	fmt.Print("Title filter: ")
	app.scanner.Scan()
	filters.Title = app.scanner.Text()

	fmt.Print("Genre filter: ")
	app.scanner.Scan()
	filters.Genre = app.scanner.Text()

	result, err := app.movieService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d movies (total: %d):\n", len(result.Data), result.Total)
	for _, movie := range result.Data {
		fmt.Printf("ID: %s, Title: %s, Duration: %s min\n",
			movie.ID, movie.Title, movie.Duration)
	}
}

func (app *TUIApp) getMovie() {
	fmt.Print("Enter movie ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	movie, err := app.movieService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", movie.ID)
	fmt.Printf("Title: %s\n", movie.Title)
	fmt.Printf("Description: %s\n", movie.Description)
	fmt.Printf("Duration: %s min\n", movie.Duration)
}

func (app *TUIApp) createMovie() {
	var req dto.CreateMovieRequest

	fmt.Print("Title: ")
	app.scanner.Scan()
	req.Title = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	fmt.Print("Duration (minutes): ")
	app.scanner.Scan()
	durationStr := app.scanner.Text()
	req.Duration = durationStr

	fmt.Print("Genre IDs (comma separated): ")
	app.scanner.Scan()
	genreIDs := app.scanner.Text()
	if genreIDs != "" {
		req.GenreIDs = strings.Split(genreIDs, ",")
	}

	id, err := app.movieService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Movie created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateMovie() {
	fmt.Print("Enter movie ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateMovieRequest
	fmt.Print("New title: ")
	app.scanner.Scan()
	req.Title = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	fmt.Print("New duration (minutes): ")
	app.scanner.Scan()
	durationStr := app.scanner.Text()
	req.Duration = durationStr

	err := app.movieService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Movie updated successfully!")
}

func (app *TUIApp) deleteMovie() {
	fmt.Print("Enter movie ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.movieService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Movie deleted successfully!")
}
