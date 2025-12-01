package handler

import (
	"net/http"

	"github.com/RandithaK/StudyBuddy_Backend/internal/server"
)

// Handler is the entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize the server (runs once)
	server.Once.Do(server.Setup)

	// Serve request
	if server.Router != nil {
		server.Router.ServeHTTP(w, r)
	} else {
		http.Error(w, "Server initialization failed", http.StatusInternalServerError)
	}
}
