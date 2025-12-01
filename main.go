package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load .env file into env variables, but ensure system environment
	// variables take precedence: for each var in .env, only set it if
	// not already present in the OS environment. This allows local .env
	// files to provide defaults while allowing system env vars to override.
	if envMap, err := godotenv.Read(); err == nil {
		for k, v := range envMap {
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
		log.Println("Loaded .env file (defaults applied; system env vars override)")
	}
	cfg := ServerConfig{
		Addr:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret"),
		Now:       time.Now,
	}

	ctx := context.Background()
	mongoURI := getEnv("MONGO_URI", "")
	store, err := NewStore(ctx, mongoURI)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}
	seedStore(store)

	// If using MongoStore, ensure we disconnect on exit.
	if ms, ok := store.(*MongoStore); ok {
		defer ms.client.Disconnect(ctx)
	}
	handler := setupRouter(store, cfg)

	log.Printf("Starting server at %s", cfg.Addr)
	if err := http.ListenAndServe(":"+cfg.Addr, handler); err != nil {
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

func setupRouter(store Store, cfg ServerConfig) http.Handler {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods(http.MethodGet)
	// Auth
	api.HandleFunc("/auth/register", RegisterHandler(store)).Methods(http.MethodPost)
	api.HandleFunc("/auth/login", LoginHandler(store, cfg)).Methods(http.MethodPost)
	// Tasks
	api.HandleFunc("/tasks", GetTasksHandler(store)).Methods(http.MethodGet)
	api.Handle("/tasks", WithAuth(cfg.JWTSecret, http.HandlerFunc(CreateTaskHandler(store)))).Methods(http.MethodPost)
	api.HandleFunc("/tasks/{id}", GetTaskHandler(store)).Methods(http.MethodGet)
	api.Handle("/tasks/{id}", WithAuth(cfg.JWTSecret, http.HandlerFunc(UpdateTaskHandler(store)))).Methods(http.MethodPut)
	api.Handle("/tasks/{id}", WithAuth(cfg.JWTSecret, http.HandlerFunc(DeleteTaskHandler(store)))).Methods(http.MethodDelete)
	// Courses
	api.HandleFunc("/courses", GetCoursesHandler(store)).Methods(http.MethodGet)
	api.Handle("/courses", WithAuth(cfg.JWTSecret, http.HandlerFunc(CreateCourseHandler(store)))).Methods(http.MethodPost)
	// Events
	api.HandleFunc("/events", GetEventsHandler(store)).Methods(http.MethodGet)
	api.Handle("/events", WithAuth(cfg.JWTSecret, http.HandlerFunc(CreateEventHandler(store)))).Methods(http.MethodPost)
	// Wrap with CORS
	return WithCORS(r)
}

func seedStore(s Store) {
	// If using MongoStore, clear collections to avoid duplicates during seeding
	if ms, ok := s.(*MongoStore); ok {
		ctx := context.Background()
		_ = ms.db.Collection("tasks").Drop(ctx)
		_ = ms.db.Collection("courses").Drop(ctx)
		_ = ms.db.Collection("events").Drop(ctx)
		_ = ms.db.Collection("users").Drop(ctx)
	}
	// Create seed courses, tasks and events similar to frontend mockData
	courses := []Course{
		{ID: "1", Name: "Biology 101", Color: "from-green-400 to-emerald-500", TotalTasks: 10, CompletedTasks: 7},
		{ID: "2", Name: "History 205", Color: "from-amber-400 to-orange-500", TotalTasks: 8, CompletedTasks: 3},
		{ID: "3", Name: "Mathematics", Color: "from-blue-400 to-cyan-500", TotalTasks: 12, CompletedTasks: 9},
		{ID: "4", Name: "Computer Science", Color: "from-purple-400 to-pink-500", TotalTasks: 15, CompletedTasks: 12},
		{ID: "5", Name: "English Literature", Color: "from-red-400 to-rose-500", TotalTasks: 6, CompletedTasks: 4},
		{ID: "6", Name: "Physics", Color: "from-indigo-400 to-violet-500", TotalTasks: 9, CompletedTasks: 5},
	}
	for _, c := range courses {
		s.CreateCourse(c)
	}

	tasks := []Task{
		{ID: "1", Title: "Chapter 5 Reading", Description: "Read and summarize chapter 5 on cell division", CourseID: "1", DueDate: "2025-11-03", DueTime: "14:00", Completed: false, HasReminder: true},
		{ID: "2", Title: "Lab Report", Description: "Complete lab report on photosynthesis experiment", CourseID: "1", DueDate: "2025-11-05", DueTime: "23:59", Completed: false, HasReminder: true},
		{ID: "3", Title: "Essay Draft", Description: "First draft of WWI causes essay", CourseID: "2", DueDate: "2025-11-02", DueTime: "17:00", Completed: false, HasReminder: true},
		{ID: "4", Title: "Problem Set 8", Description: "Calculus problems on derivatives", CourseID: "3", DueDate: "2025-11-04", DueTime: "10:00", Completed: false, HasReminder: false},
		{ID: "5", Title: "Code Assignment", Description: "Implement binary search tree", CourseID: "4", DueDate: "2025-11-06", DueTime: "23:59", Completed: false, HasReminder: true},
		{ID: "6", Title: "Poetry Analysis", Description: "Analyze Robert Frost poems", CourseID: "5", DueDate: "2025-11-01", DueTime: "16:00", Completed: true, HasReminder: false},
	}
	for _, t := range tasks {
		s.CreateTask(t)
	}

	events := []Event{
		{ID: "1", Title: "Biology Lecture", CourseID: "1", Date: "2025-11-01", StartTime: "09:00", EndTime: "10:30", Type: "class"},
		{ID: "2", Title: "Study Session", CourseID: "3", Date: "2025-11-01", StartTime: "14:00", EndTime: "16:00", Type: "study"},
		{ID: "3", Title: "History Seminar", CourseID: "2", Date: "2025-11-01", StartTime: "11:00", EndTime: "12:30", Type: "class"},
		{ID: "4", Title: "Math Midterm", CourseID: "3", Date: "2025-11-04", StartTime: "13:00", EndTime: "15:00", Type: "exam"},
		{ID: "5", Title: "CS Lab", CourseID: "4", Date: "2025-11-02", StartTime: "15:00", EndTime: "17:00", Type: "class"},
	}
	for _, e := range events {
		s.CreateEvent(e)
	}

	// Seed a demo user (email: test@example.com, password: password)
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err == nil {
		user := User{ID: uuid.New().String(), Name: "Test User", Email: "test@example.com", Password: string(hash)}
		s.CreateUser(user)
	}
}

// duplicate functions removed
