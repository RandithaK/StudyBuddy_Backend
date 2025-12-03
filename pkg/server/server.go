package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/RandithaK/StudyBuddy_Backend/graph"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/auth"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/email"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/models"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/store"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/worker"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	Router *mux.Router
	St     store.Store
	Once   sync.Once
)

// Setup initializes the database and router.
// It is public (capitalized) so it can be called from main.go and api/index.go
func Setup() {
	// Load .env if present
	if envMap, err := godotenv.Read(); err == nil {
		for k, v := range envMap {
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}

	mongoURI := GetEnv("MONGO_URI", "")
	if mongoURI == "" {
		log.Println("Warning: MONGO_URI is empty")
	}

	ctx := context.Background()
	var err error

	if St == nil {
		St, err = store.NewStore(ctx, mongoURI)
		if err != nil {
			log.Printf("failed to create store: %v", err)
			return
		}

		// Optional: Seed data (skip on Vercel production to avoid cold start delays)
		if os.Getenv("VERCEL") != "1" {
			SeedStore(St)
		}
	}

	// Start Worker (Only for local dev usually, or check flags)
	if os.Getenv("VERCEL") != "1" {
		w := worker.NewWorker(St)
		w.Start()
	}

	if Router == nil {
		Router = SetupRouter(St)
	}
}

func SetupRouter(s store.Store) *mux.Router {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Store: s}}))

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Use(authMiddleware)

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	r.HandleFunc("/verify-email", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}

		user, err := s.GetUserByVerificationToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusBadRequest)
			return
		}

		if err := s.MarkUserVerified(user.ID); err != nil {
			http.Error(w, "Failed to verify email", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
            <html>
                <head><title>Email Verified</title></head>
                <body style="font-family: sans-serif; text-align: center; padding: 50px;">
                    <h1 style="color: green;">Email Verified!</h1>
                    <p>Your email has been successfully verified.</p>
                    <p>You can now <a href="studybuddy://login">return to the app</a> and login.</p>
                </body>
            </html>
        `))
	}).Methods(http.MethodGet)

	// Client-triggered email fallback (called by app background fetch)
	r.HandleFunc("/api/notifications/check-email-fallback", func(w http.ResponseWriter, r *http.Request) {
		// Get UserID from context (set by authMiddleware)
		userID, ok := r.Context().Value(auth.UserIDKey).(string)
		if !ok || userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get unread notifications older than 1 hour for this user
		notifications, err := s.GetUnreadNotificationsOlderThanForUser(userID, "1h")
		if err != nil {
			log.Printf("Error getting unread notifications for user %s: %v", userID, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		user, err := s.GetUser(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if !user.IsVerified {
			// Just mark as emailed to avoid repeated checks
			for _, n := range notifications {
				s.MarkNotificationAsEmailed(n.ID)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("User not verified, skipped emails"))
			return
		}

		count := 0
		for _, n := range notifications {
			err = email.SendNotificationEmail(user.Email, "You have an unread notification", n.Message)
			if err != nil {
				log.Printf("Error sending email to %s: %v", user.Email, err)
				continue
			}
			s.MarkNotificationAsEmailed(n.ID)
			count++
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Processed %d notifications", count)
	}).Methods(http.MethodPost)

	// GraphQL playground and handlers
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", srv)
	// Support Vercel's /api/* route prefix in production deployments
	// (e.g. https://<host>/api/query). This ensures requests made to
	// '/api/query' are handled properly when the Vercel router passes through the path.
	r.Handle("/api/query", srv)

	return r
}

func GetEnv(key, fallback string) string {
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

func SeedStore(s store.Store) {
	// (Keep your seed logic here exactly as it was)
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		ID:       "test-user-id",
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hash),
	}
	if _, exists := s.GetUserByEmail(user.Email); !exists {
		s.CreateUser(user)
	}
	// ... rest of your seed logic ...
	// Seed sample courses and tasks (mirrors the previous implementation in main)
	userID := "test-user-id"

	courses := []models.Course{
		{ID: "course-1", Name: "CS50", Color: "#f59e0b", UserID: userID},
		{ID: "course-2", Name: "Biology 101", Color: "#10b981", UserID: userID},
		{ID: "course-3", Name: "Math 201", Color: "#3b82f6", UserID: userID},
	}

	for _, course := range courses {
		s.CreateCourse(course)
	}

	// Seed sample tasks linked to courses
	tasks := []models.Task{
		{
			ID:          "task-1",
			Title:       "Problem Set 1",
			Description: "Complete problem set 1",
			CourseID:    "course-1",
			UserID:      userID,
			DueDate:     "2025-12-10",
			DueTime:     "23:59",
			Completed:   false,
			HasReminder: true,
		},
		{
			ID:          "task-2",
			Title:       "Problem Set 2",
			Description: "Complete problem set 2",
			CourseID:    "course-1",
			UserID:      userID,
			DueDate:     "2025-12-12",
			DueTime:     "23:59",
			Completed:   true,
			HasReminder: false,
		},
		{
			ID:          "task-3",
			Title:       "Lab Report",
			Description: "Write lab report",
			CourseID:    "course-2",
			UserID:      userID,
			DueDate:     "2025-12-08",
			DueTime:     "17:00",
			Completed:   false,
			HasReminder: true,
		},
		{
			ID:          "task-4",
			Title:       "Essay",
			Description: "Write essay",
			CourseID:    "course-2",
			UserID:      userID,
			DueDate:     "2025-12-15",
			DueTime:     "23:59",
			Completed:   false,
			HasReminder: false,
		},
		{
			ID:          "task-5",
			Title:       "Calculus Practice",
			Description: "Practice derivatives",
			CourseID:    "course-3",
			UserID:      userID,
			DueDate:     "2025-12-05",
			DueTime:     "14:00",
			Completed:   false,
			HasReminder: true,
		},
	}

	for _, task := range tasks {
		s.CreateTask(task)
	}
}
