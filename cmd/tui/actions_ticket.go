package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) manageTickets() {
	for {
		fmt.Println("\n--- Ticket Management ---")
		fmt.Println("1. List tickets")
		fmt.Println("2. Get ticket by ID")
		fmt.Println("3. Create ticket")
		fmt.Println("4. Update ticket Status")
		fmt.Println("5. Delete ticket")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listTickets()
		case "2":
			app.getTicket()
		case "3":
			app.createTicket()
		case "4":
			app.updateStatus()
		case "5":
			app.deleteTicket()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listTickets() {
	page, limit := app.getPaginationParams()

	var filters dto.TicketFilters
	fmt.Print("Movie Show ID: ")
	app.scanner.Scan()
	filters.MovieShowID = app.scanner.Text()

	fmt.Print("Statuses (comma separated): ")
	app.scanner.Scan()
	Statuses := app.scanner.Text()
	if Statuses != "" {
		filters.Status = strings.Split(Statuses, ",")
	}

	result, err := app.ticketService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d tickets (total: %d):\n", len(result.Data), result.Total)
	for _, ticket := range result.Data {
		fmt.Printf("ID: %s, Movie Show: %s, Seat: %s, Status: %s\n",
			ticket.ID, ticket.MovieShowID, ticket.SeatID, ticket.Status)
	}
}

func (app *TUIApp) getTicket() {
	fmt.Print("Enter ticket ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	ticket, err := app.ticketService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", ticket.ID)
	fmt.Printf("Movie Show ID: %s\n", ticket.MovieShowID)
	fmt.Printf("Seat ID: %s\n", ticket.SeatID)
	fmt.Printf("User ID: %s\n", ticket.UserID)
	fmt.Printf("Status: %s\n", ticket.Status)
	fmt.Printf("Price: %.2f\n", ticket.Price)
}

func (app *TUIApp) createTicket() {
	var req dto.CreateTicketRequest

	fmt.Print("Movie Show ID: ")
	app.scanner.Scan()
	movieShowID := app.scanner.Text()

	fmt.Print("Seat ID: ")
	app.scanner.Scan()
	req.SeatID = app.scanner.Text()

	fmt.Print("Price: ")
	app.scanner.Scan()
	priceStr := app.scanner.Text()
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		req.Price = price
	}

	id, err := app.ticketService.Create(context.Background(), req, movieShowID)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Ticket created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateStatus() {
	fmt.Print("Enter ticket ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateStatusRequest
	fmt.Print("New Status: ")
	app.scanner.Scan()
	req.Status = app.scanner.Text()

	err := app.ticketService.UpdateStatus(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Ticket Status updated successfully!")
}

func (app *TUIApp) deleteTicket() {
	fmt.Print("Enter ticket ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.ticketService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Ticket deleted successfully!")
}
