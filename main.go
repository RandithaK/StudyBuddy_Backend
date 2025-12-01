package main

import (
	"log"
	"net/http"

	"github.com/RandithaK/StudyBuddy/backend/internal/server"
)

func main() {
	server.Setup()

	port := server.GetEnv("PORT", "8080")
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	if err := http.ListenAndServe(":"+port, server.Router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
