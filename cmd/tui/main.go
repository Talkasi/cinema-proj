package main

import "cw/internal/tui/client"

const defaultAPIBaseURL = "http://localhost:8080"

func main() {
	apiClient := client.NewAPIClient(defaultAPIBaseURL)
	app := NewTUIApp(apiClient)
	app.Run()
}
