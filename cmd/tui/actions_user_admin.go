package main

import (
	"context"
	"fmt"
	"strings"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) manageUsers() {
	for {
		fmt.Println("\n--- User Management ---")
		fmt.Println("1. List users")
		fmt.Println("2. Get user by ID")
		fmt.Println("3. Update user")
		fmt.Println("4. Update admin Status")
		fmt.Println("5. Delete user")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listUsers()
		case "2":
			app.getUser()
		case "3":
			app.updateUser()
		case "4":
			app.updateAdminStatus()
		case "5":
			app.deleteUser()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listUsers() {
	page, limit := app.getPaginationParams()

	var filters dto.UserFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	fmt.Print("Email filter: ")
	app.scanner.Scan()
	filters.Email = app.scanner.Text()

	result, err := app.userService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("\nFound %d users (total: %d):\n", len(result.Data), result.Total)
	for _, user := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Email: %s, Admin: %t\n",
			user.ID, user.Name, user.Email, user.IsAdmin)
	}
}

func (app *TUIApp) getUser() {
	fmt.Print("Enter user ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	user, err := app.userService.GetByID(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Is Admin: %t\n", user.IsAdmin)
}

func (app *TUIApp) updateUser() {
	fmt.Print("Enter user ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateUserRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New email: ")
	app.scanner.Scan()
	req.Email = app.scanner.Text()

	err := app.userService.Update(context.Background(), id, req)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("User updated successfully!")
}

func (app *TUIApp) updateAdminStatus() {
	fmt.Print("Enter user ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	fmt.Print("Is admin (true/false): ")
	app.scanner.Scan()
	isAdminStr := app.scanner.Text()
	isAdmin := strings.ToLower(isAdminStr) == "true"

	err := app.userService.UpdateAdminStatus(context.Background(), id, isAdmin)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("Admin Status updated successfully!")
}

func (app *TUIApp) deleteUser() {
	fmt.Print("Enter user ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.userService.Delete(context.Background(), id)
	if err != nil {
		app.printError(err)
		return
	}

	fmt.Println("User deleted successfully!")
}
