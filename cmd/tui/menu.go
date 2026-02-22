package main

import (
	"fmt"
	"os"
)

func (app *TUIApp) showAuthMenu() {
	fmt.Println("\n--- Authentication ---")
	fmt.Println("1. Login")
	fmt.Println("2. Register")
	fmt.Println("3. Exit")
	fmt.Print("Choose option: ")

	app.scanner.Scan()
	choice := app.scanner.Text()

	switch choice {
	case "1":
		app.login()
	case "2":
		app.register()
	case "3":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid option!")
	}
}

func (app *TUIApp) showMainMenu() {
	app.displayMainMenuOptions()
	choice := app.getMenuChoice()
	app.processMenuChoice(choice)
}

func (app *TUIApp) displayMainMenuOptions() {
	fmt.Println("\n--- Main Menu ---")
	fmt.Println("1. Manage Genres")
	fmt.Println("2. Manage Movies")
	fmt.Println("3. Manage Movie Shows")
	fmt.Println("4. Manage Tickets")
	fmt.Println("5. Manage Reviews")
	fmt.Println("6. Manage Halls")
	fmt.Println("7. Manage Screen Types")
	fmt.Println("8. Manage Seat Types")
	fmt.Println("9. Manage Seats")
	if app.isAdmin {
		fmt.Println("10. Manage Users")
	}
	fmt.Println("0. Logout")
	fmt.Print("Choose option: ")
}

func (app *TUIApp) getMenuChoice() string {
	app.scanner.Scan()
	return app.scanner.Text()
}

func (app *TUIApp) processMenuChoice(choice string) {
	choiceKey, ok := app.resolveMainMenuChoice(choice)
	if !ok {
		fmt.Println("Invalid option!")
		return
	}

	action := app.getActionForChoice(choiceKey)
	if action != nil {
		action()
	} else {
		fmt.Println("Invalid option!")
	}
}

func (app *TUIApp) resolveMainMenuChoice(choice string) (string, bool) {
	if choice == "10" && !app.isAdmin {
		return "", false
	}

	if app.getActionForChoice(choice) == nil {
		return "", false
	}

	return choice, true
}

func (app *TUIApp) mainMenuActions() map[string]func() {
	return map[string]func(){
		"1":  app.manageGenres,
		"2":  app.manageMovies,
		"3":  app.manageMovieShows,
		"4":  app.manageTickets,
		"5":  app.manageReviews,
		"6":  app.manageHalls,
		"7":  app.manageScreenTypes,
		"8":  app.manageSeatTypes,
		"9":  app.manageSeats,
		"10": func() { app.handleUserManagementOption() },
		"0":  app.logout,
	}
}

func (app *TUIApp) getActionForChoice(choice string) func() {
	return app.mainMenuActions()[choice]
}

func (app *TUIApp) handleUserManagementOption() {
	if app.isAdmin {
		app.manageUsers()
	} else {
		fmt.Println("Invalid option!")
	}
}
