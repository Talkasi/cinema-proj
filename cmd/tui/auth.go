package main

import (
	"context"
	"fmt"

	dto "cw/internal/dto/models"
)

func (app *TUIApp) login() {
	var email, password string

	fmt.Print("Email: ")
	app.scanner.Scan()
	email = app.scanner.Text()

	fmt.Print("PasswordHash: ")
	app.scanner.Scan()
	password = app.scanner.Text()

	token, err := app.authService.Login(context.Background(), email, password)
	if err != nil {
		app.printErrorWithPrefix("Login failed", err)
		return
	}

	app.token = token
	app.isAuthenticated = true
	app.isAdmin = true
	fmt.Println("Login successful!")
}

func (app *TUIApp) register() {
	var req dto.CreateUserRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Email: ")
	app.scanner.Scan()
	req.Email = app.scanner.Text()

	fmt.Print("PasswordHash: ")
	app.scanner.Scan()
	req.PasswordHash = app.scanner.Text()

	id, err := app.authService.Register(context.Background(), req)
	if err != nil {
		app.printErrorWithPrefix("Registration failed", err)
		return
	}

	fmt.Printf("Registration successful! User ID: %s\n", id)
}

func (app *TUIApp) logout() {
	app.isAuthenticated = false
	app.isAdmin = false
	app.token = ""
	fmt.Println("Logged out successfully!")
}
