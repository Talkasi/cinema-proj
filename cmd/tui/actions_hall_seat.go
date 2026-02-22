package main

import (
	"context"
	"fmt"
	"strconv"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) manageHalls() {
	for {
		fmt.Println("\n--- Hall Management ---")
		fmt.Println("1. List halls")
		fmt.Println("2. Get hall by ID")
		fmt.Println("3. Create hall")
		fmt.Println("4. Update hall")
		fmt.Println("5. Delete hall")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listHalls()
		case "2":
			app.getHall()
		case "3":
			app.createHall()
		case "4":
			app.updateHall()
		case "5":
			app.deleteHall()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listHalls() {
	page, limit := app.getPaginationParams()

	var filters dto.HallFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	fmt.Print("Screen Type ID: ")
	app.scanner.Scan()
	filters.ScreenTypeID = app.scanner.Text()

	result, err := app.hallService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d halls (total: %d):\n", len(result.Data), result.Total)
	for _, hall := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Screen Type: %s\n",
			hall.ID, hall.Name, hall.ScreenTypeID)
	}
}

func (app *TUIApp) getHall() {
	fmt.Print("Enter hall ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	hall, err := app.hallService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", hall.ID)
	fmt.Printf("Name: %s\n", hall.Name)
	fmt.Printf("Screen Type ID: %s\n", hall.ScreenTypeID)
	fmt.Printf("Description: %s\n", hall.Description)
}

func (app *TUIApp) createHall() {
	var req dto.CreateHallRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Screen Type ID: ")
	app.scanner.Scan()
	req.ScreenTypeID = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.hallService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Hall created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateHall() {
	fmt.Print("Enter hall ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateHallRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New screen type ID: ")
	app.scanner.Scan()
	req.ScreenTypeID = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.hallService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Hall updated successfully!")
}

func (app *TUIApp) deleteHall() {
	fmt.Print("Enter hall ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.hallService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Hall deleted successfully!")
}

func (app *TUIApp) manageScreenTypes() {
	for {
		fmt.Println("\n--- Screen Type Management ---")
		fmt.Println("1. List screen types")
		fmt.Println("2. Get screen type by ID")
		fmt.Println("3. Create screen type")
		fmt.Println("4. Update screen type")
		fmt.Println("5. Delete screen type")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listScreenTypes()
		case "2":
			app.getScreenType()
		case "3":
			app.createScreenType()
		case "4":
			app.updateScreenType()
		case "5":
			app.deleteScreenType()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listScreenTypes() {
	page, limit := app.getPaginationParams()

	var filters dto.ScreenTypeFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	result, err := app.screenTypeService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d screen types (total: %d):\n", len(result.Data), result.Total)
	for _, st := range result.Data {
		fmt.Printf("ID: %s, Name: %s\n", st.ID, st.Name)
	}
}

func (app *TUIApp) getScreenType() {
	fmt.Print("Enter screen type ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	st, err := app.screenTypeService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", st.ID)
	fmt.Printf("Name: %s\n", st.Name)
	fmt.Printf("Description: %s\n", st.Description)
}

func (app *TUIApp) createScreenType() {
	var req dto.CreateScreenTypeRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.screenTypeService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Screen type created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateScreenType() {
	fmt.Print("Enter screen type ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateScreenTypeRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.screenTypeService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Screen type updated successfully!")
}

func (app *TUIApp) deleteScreenType() {
	fmt.Print("Enter screen type ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.screenTypeService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Screen type deleted successfully!")
}

func (app *TUIApp) manageSeatTypes() {
	for {
		fmt.Println("\n--- Seat Type Management ---")
		fmt.Println("1. List seat types")
		fmt.Println("2. Get seat type by ID")
		fmt.Println("3. Create seat type")
		fmt.Println("4. Update seat type")
		fmt.Println("5. Delete seat type")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listSeatTypes()
		case "2":
			app.getSeatType()
		case "3":
			app.createSeatType()
		case "4":
			app.updateSeatType()
		case "5":
			app.deleteSeatType()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listSeatTypes() {
	page, limit := app.getPaginationParams()

	var filters dto.SeatTypeFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	result, err := app.seatTypeService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d seat types (total: %d):\n", len(result.Data), result.Total)
	for _, st := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Price Multiplier: %.2f\n",
			st.ID, st.Name, st.PriceModifier)
	}
}

func (app *TUIApp) getSeatType() {
	fmt.Print("Enter seat type ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	st, err := app.seatTypeService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", st.ID)
	fmt.Printf("Name: %s\n", st.Name)
	fmt.Printf("Price Multiplier: %.2f\n", st.PriceModifier)
	fmt.Printf("Description: %s\n", st.Description)
}

func (app *TUIApp) createSeatType() {
	var req dto.CreateSeatTypeRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Price Multiplier: ")
	app.scanner.Scan()
	multiplierStr := app.scanner.Text()
	if multiplier, err := strconv.ParseFloat(multiplierStr, 64); err == nil {
		req.PriceModifier = multiplier
	}

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.seatTypeService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Seat type created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateSeatType() {
	fmt.Print("Enter seat type ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateSeatTypeRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New price multiplier: ")
	app.scanner.Scan()
	multiplierStr := app.scanner.Text()
	if multiplier, err := strconv.ParseFloat(multiplierStr, 64); err == nil {
		req.PriceModifier = multiplier
	}

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.seatTypeService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Seat type updated successfully!")
}

func (app *TUIApp) deleteSeatType() {
	fmt.Print("Enter seat type ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.seatTypeService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Seat type deleted successfully!")
}

func (app *TUIApp) manageSeats() {
	for {
		fmt.Println("\n--- Seat Management ---")
		fmt.Println("1. List seats by hall")
		fmt.Println("2. Get seat by ID")
		fmt.Println("3. Create seat")
		fmt.Println("4. Update seat")
		fmt.Println("5. Delete seat")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listSeatsByHall()
		case "2":
			app.getSeat()
		case "3":
			app.createSeat()
		case "4":
			app.updateSeat()
		case "5":
			app.deleteSeat()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listSeatsByHall() {
	fmt.Print("Enter hall ID: ")
	app.scanner.Scan()
	hallID := app.scanner.Text()

	page, limit := app.getPaginationParams()

	var filters dto.SeatFilters
	fmt.Print("Seat Type ID: ")
	app.scanner.Scan()
	filters.SeatTypeID = app.scanner.Text()

	result, err := app.seatService.GetByHall(context.Background(), hallID, filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d seats (total: %d):\n", len(result.Data), result.Total)
	for _, seat := range result.Data {
		fmt.Printf("ID: %s, Row: %d, Number: %d, Type: %s\n",
			seat.ID, seat.RowNumber, seat.SeatNumber, seat.SeatTypeID)
	}
}

func (app *TUIApp) getSeat() {
	fmt.Print("Enter seat ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	seat, err := app.seatService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", seat.ID)
	fmt.Printf("Hall ID: %s\n", seat.HallID)
	fmt.Printf("Seat Type ID: %s\n", seat.SeatTypeID)
	fmt.Printf("Row: %d\n", seat.RowNumber)
	fmt.Printf("Number: %d\n", seat.SeatNumber)
}

func (app *TUIApp) createSeat() {
	var req dto.CreateSeatRequest

	fmt.Print("Hall ID: ")
	app.scanner.Scan()
	req.HallID = app.scanner.Text()

	fmt.Print("Seat Type ID: ")
	app.scanner.Scan()
	req.SeatTypeID = app.scanner.Text()

	fmt.Print("Row Number: ")
	app.scanner.Scan()
	rowStr := app.scanner.Text()
	if row, err := strconv.Atoi(rowStr); err == nil {
		req.RowNumber = row
	}

	fmt.Print("Seat Number: ")
	app.scanner.Scan()
	seatStr := app.scanner.Text()
	if seat, err := strconv.Atoi(seatStr); err == nil {
		req.SeatNumber = seat
	}

	id, err := app.seatService.Create(context.Background(), req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("Seat created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateSeat() {
	fmt.Print("Enter seat ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateSeatRequest
	fmt.Print("New seat type ID: ")
	app.scanner.Scan()
	req.SeatTypeID = app.scanner.Text()

	fmt.Print("New row number: ")
	app.scanner.Scan()
	rowStr := app.scanner.Text()
	if row, err := strconv.Atoi(rowStr); err == nil {
		req.RowNumber = row
	}

	fmt.Print("New seat number: ")
	app.scanner.Scan()
	seatStr := app.scanner.Text()
	if seat, err := strconv.Atoi(seatStr); err == nil {
		req.SeatNumber = seat
	}

	err := app.seatService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Seat updated successfully!")
}

func (app *TUIApp) deleteSeat() {
	fmt.Print("Enter seat ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.seatService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Seat deleted successfully!")
}
