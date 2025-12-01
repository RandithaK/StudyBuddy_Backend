package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "strings"
    "sync"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    "github.com/RandithaK/StudyBuddy/backend/graph"
    "github.com/RandithaK/StudyBuddy/backend/internal/auth"
    "github.com/RandithaK/StudyBuddy/backend/internal/models"
    "github.com/RandithaK/StudyBuddy/backend/internal/store"
    "github.com/RandithaK/StudyBuddy/backend/internal/worker"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "golang.org/x/crypto/bcrypt"
)

var (
    router *mux.Router
    st     store.Store
    once   sync.Once
)

// Main is only used for LOCAL development (go run main.go)
func main() {
    // Load env locally
    if envMap, err := godotenv.Read(); err == nil {
        for k, v := range envMap {
            if os.Getenv(k) == "" {
                os.Setenv(k, v)
            }
        }
        log.Println("Loaded .env file")
    }

    port := getEnv("PORT", "8080")
    
    // Initialize app
    setup()

    log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
    if err := http.ListenAndServe(":"+port, router); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

// Handler is the entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
    // Ensure setup runs once per cold start
    once.Do(setup)
    
    // Delegate to the router
    router.ServeHTTP(w, r)
}

// setup initializes the DB and Router
func setup() {
    mongoURI := getEnv("MONGO_URI", "")
    if mongoURI == "" {
        log.Println("Warning: MONGO_URI is empty")
    }

    ctx := context.Background()
    var err error
    
    // Connect to Store
    // In Vercel, we reuse the connection if the variable 'st' is already set from a previous warm request
    if st == nil {
        st, err = store.NewStore(ctx, mongoURI)
        if err != nil {
            // In a handler, we can't Fatal, so we log. 
            // In a real app, you might want to handle this gracefully in the HTTP response.
            log.Printf("failed to create store: %v", err)
            return 
        }
        
        // Seed data (Optional: Be careful seeding on every cold start in production)
        // You might want to remove this for production or check a flag
        if os.Getenv("VERCEL") != "1" {
            seedStore(st)
        }
    }

    // Start Worker
    // WARNING: On Vercel, background goroutines freeze when the request ends.
    // This worker will only run while a user is actively making a request.
    // Ideally, disable this on Vercel and use Vercel Cron or an external job.
    if os.Getenv("VERCEL") != "1" {
        w := worker.NewWorker(st)
        w.Start()
    }

    // Initialize Router
    if router == nil {
        router = SetupRouter(st)
    }
}

func SetupRouter(s store.Store, port ...string) *mux.Router {
    // Note: port arg is unused in logic but kept for signature compatibility if needed, 
    // though removed from signature here for cleanliness.
    
    srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Store: s}}))

    r := mux.NewRouter()
    r.Use(loggingMiddleware)
    r.Use(authMiddleware)

    // Health check
    r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    }).Methods(http.MethodGet)

    // Email Verification
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

    // GraphQL
    r.Handle("/", playground.Handler("GraphQL playground", "/query"))
    r.Handle("/query", srv)

    return r
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

    // Seed sample courses and tasks
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