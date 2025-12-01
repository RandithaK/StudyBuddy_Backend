package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondErr(w http.ResponseWriter, code int, msg string) {
	respondJSON(w, code, map[string]string{"error": msg})
}

// Auth Handlers
func RegisterHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		if req.Name == "" || req.Email == "" || req.Password == "" {
			respondErr(w, http.StatusBadRequest, "name, email and password required")
			return
		}
		if _, ok := s.GetUserByEmail(req.Email); ok {
			respondErr(w, http.StatusBadRequest, "user already exists")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			respondErr(w, http.StatusInternalServerError, "unable to hash password")
			return
		}
		user := User{
			ID:       uuid.New().String(),
			Name:     req.Name,
			Email:    req.Email,
			Password: string(hash),
		}
		s.CreateUser(user)
		respondJSON(w, http.StatusCreated, map[string]interface{}{"id": user.ID, "name": user.Name, "email": user.Email})
	}
}

func LoginHandler(s Store, cfg ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		user, ok := s.GetUserByEmail(req.Email)
		if !ok {
			respondErr(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			respondErr(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		claims := jwt.MapClaims{
			"userId": user.ID,
			"email":  user.Email,
			"name":   user.Name,
			"exp":    time.Now().Add(24 * time.Hour).Unix(),
		}
		tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tkn.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			respondErr(w, http.StatusInternalServerError, "could not create token")
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"token": signed})
	}
}

// Task handlers
func GetTasksHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks := s.GetTasks()
		respondJSON(w, http.StatusOK, tasks)
	}
}

func CreateTaskHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		if t.ID == "" {
			t.ID = uuid.New().String()
		}
		if t.Title == "" {
			respondErr(w, http.StatusBadRequest, "title required")
			return
		}
		task := s.CreateTask(t)
		respondJSON(w, http.StatusCreated, task)
	}
}

func GetTaskHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		t, err := s.GetTask(id)
		if err != nil {
			respondErr(w, http.StatusNotFound, "task not found")
			return
		}
		respondJSON(w, http.StatusOK, t)
	}
}

func UpdateTaskHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		updated, err := s.UpdateTask(id, t)
		if err != nil {
			respondErr(w, http.StatusNotFound, "task not found")
			return
		}
		respondJSON(w, http.StatusOK, updated)
	}
}

func DeleteTaskHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if err := s.DeleteTask(id); err != nil {
			respondErr(w, http.StatusNotFound, "task not found")
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"id": id})
	}
}

// Courses handlers
func GetCoursesHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courses := s.GetCourses()
		respondJSON(w, http.StatusOK, courses)
	}
}

func CreateCourseHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Course
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		if c.ID == "" {
			c.ID = uuid.New().String()
		}
		course := s.CreateCourse(c)
		respondJSON(w, http.StatusCreated, course)
	}
}

// Events handlers
func GetEventsHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events := s.GetEvents()
		respondJSON(w, http.StatusOK, events)
	}
}

func CreateEventHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		if e.ID == "" {
			e.ID = uuid.New().String()
		}
		evt := s.CreateEvent(e)
		respondJSON(w, http.StatusCreated, evt)
	}
}
