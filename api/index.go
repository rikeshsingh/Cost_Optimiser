package handler

import (
	"fmt"
	"net/http"
)

// Handler is the entry point Vercel uses for serverless functions
func Handler(w http.ResponseWriter, r *http.Request) {
	// Your custom backend logic goes here
	fmt.Fprintf(w, "Hello from the Cost Optimiser Go Backend!")
}
