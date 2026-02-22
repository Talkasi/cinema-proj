package main

import (
	"context"
	"fmt"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) manageGenres() {
	for {
		fmt.Println("\n--- Genre Management ---")
		fmt.Println("1. List genres")
		fmt.Println("2. Get genre by ID")
		fmt.Println("3. Create genre")
		fmt.Println("4. Update genre")
		fmt.Println("5. Delete genre")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listGenres()
		case "2":
			app.getGenre()
		case "3":
			app.createGenre()
		case "4":
			app.updateGenre()
		case "5":
			app.deleteGenre()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listGenres() {
	page, limit := app.getPaginationParams()

	var filters dto.GenreFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	fmt.Print("Description filter: ")
	app.scanner.Scan()
	filters.Description = app.scanner.Text()

	result, err := app.genreService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d genres (total: %d):\n", len(result.Data), result.Total)
	for _, genre := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Description: %s\n",
			genre.ID, genre.Name, genre.Description)
	}
}

func (app *TUIApp) getGenre() {
	fmt.Print("Enter genre ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	genre, err := app.genreService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", genre.ID)
	fmt.Printf("Name: %s\n", genre.Name)
	fmt.Printf("Description: %s\n", genre.Description)
}

func (app *TUIApp) createGenre() {
	var req dto.CreateGenreRequest

	fmt.Print("Genre name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.genreService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Genre created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateGenre() {
	fmt.Print("Enter genre ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateGenreRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.genreService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Genre updated successfully!")
}

func (app *TUIApp) deleteGenre() {
	fmt.Print("Enter genre ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.genreService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Genre deleted successfully!")
}
