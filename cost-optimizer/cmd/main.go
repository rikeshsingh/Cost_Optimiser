package main

import (
	"log"
	"net/http"

	"://github.com"
)

// CORS middleware to allow frontend requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handler is the entry point that Vercel invokes for every serverless request.
// It bypasses the traditional main() loop during production executions.
func Handler(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/cost", api.GetCostHandler)
	mux.HandleFunc("/ec2", api.GetEC2InstancesHandler)
	mux.HandleFunc("/services", api.GetAllServicesHandler)
	mux.HandleFunc("/security", api.GetSecurityHandler)
	mux.HandleFunc("/security-details", api.GetSecurityDetailsHandler)

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// Apply CORS middleware layers
	handler := corsMiddleware(mux)

	// Delegate processing context to the request router
	handler.ServeHTTP(w, r)
}

// main handles local testing execution instances.
func main() {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/cost", api.GetCostHandler)
	mux.HandleFunc("/ec2", api.GetEC2InstancesHandler)
	mux.HandleFunc("/services", api.GetAllServicesHandler)
	mux.HandleFunc("/security", api.GetSecurityHandler)
	mux.HandleFunc("/security-details", api.GetSecurityDetailsHandler)

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// Apply CORS middleware layers
	handler := corsMiddleware(mux)

	log.Println("Server running locally on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
