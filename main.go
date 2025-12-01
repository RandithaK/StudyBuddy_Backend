package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/RandithaK/StudyBuddy/backend/graph"
	"github.com/RandithaK/StudyBuddy/backend/internal/auth"
	"github.com/RandithaK/StudyBuddy/backend/internal/models"
	"github.com/RandithaK/StudyBuddy/backend/internal/store"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if envMap, err := godotenv.Read(); err == nil {
		for k, v := range envMap {
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
		log.Println("Loaded .env file")
	}

	port := getEnv("PORT", "8080")
	mongoURI := getEnv("MONGO_URI", "")

	ctx := context.Background()
	s, err := store.NewStore(ctx, mongoURI)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}

	// Seed data if needed (optional, but good for dev)
	// seedStore(s)

	if ms, ok := s.(*store.MongoStore); ok {
		// Access client via reflection or just ignore disconnect for now as it's main
		// Ideally MongoStore should expose Close()
		_ = ms
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Store: s}}))

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Use(authMiddleware)

	// Health check
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	// GraphQL
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) == 2 {
			tokenStr := bearerToken[1]
			claims, err := auth.ValidateToken(tokenStr)
			if err == nil {
				ctx := context.WithValue(r.Context(), auth.UserIDKey, claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Seed helper (simplified)
func seedStore(s store.Store) {
	// Only seed if empty? Or just ensure test user exists
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		ID:       "test-user-id",
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hash),
	}
	// Check if exists first to avoid error or duplicates
	if _, exists := s.GetUserByEmail(user.Email); !exists {
		s.CreateUser(user)
	}
}
