package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (app *TUIApp) readLine() string {
	app.scanner.Scan()
	return app.scanner.Text()
}

func (app *TUIApp) promptString(prompt string) string {
	fmt.Print(prompt)
	return app.readLine()
}

func (app *TUIApp) printError(err error) {
	fmt.Printf("Error: %v\n", err)
}

func (app *TUIApp) printErrorWithPrefix(prefix string, err error) {
	fmt.Printf("%s: %v\n", prefix, err)
}

func (app *TUIApp) printSuccess(message string) {
	fmt.Println(message)
}

func parseOptionalInt(input string, fallback int) int {
	input = strings.TrimSpace(input)
	if input == "" {
		return fallback
	}
	if value, err := strconv.Atoi(input); err == nil {
		return value
	}
	return fallback
}

func parseOptionalFloat(input string, fallback float64) float64 {
	input = strings.TrimSpace(input)
	if input == "" {
		return fallback
	}
	if value, err := strconv.ParseFloat(input, 64); err == nil {
		return value
	}
	return fallback
}

func (app *TUIApp) getPaginationParams() (int, int) {
	page := parseOptionalInt(app.promptString("Page (default 1): "), 1)
	limit := parseOptionalInt(app.promptString("Limit (default 20): "), 20)
	return page, limit
}
