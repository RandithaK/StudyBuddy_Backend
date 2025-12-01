package models

import "time"

// Task mirrors the frontend Task model
type Task struct {
	ID          string `json:"id" bson:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CourseID    string `json:"courseId"`
	UserID      string `json:"userId" bson:"userId"`
	DueDate     string `json:"dueDate"`
	DueTime     string `json:"dueTime"`
	Completed   bool   `json:"completed"`
	HasReminder bool   `json:"hasReminder"`
}

// Course mirrors the frontend Course model
type Course struct {
	ID             string `json:"id" bson:"id"`
	Name           string `json:"name"`
	Color          string `json:"color"`
	UserID         string `json:"userId" bson:"userId"`
	TotalTasks     int    `json:"totalTasks"`
	CompletedTasks int    `json:"completedTasks"`
}

// Event mirrors the frontend Event model
type Event struct {
	ID        string `json:"id" bson:"id"`
	Title     string `json:"title"`
	CourseID  string `json:"courseId"`
	UserID    string `json:"userId" bson:"userId"`
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Type      string `json:"type"`
}

// User model for authentication
type User struct {
	ID       string `json:"id" bson:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"` // hashed password
}

// Claims used for jwt
// Claims are defined in handlers to avoid coupling this package to JWT here.

// Server config
type ServerConfig struct {
	Addr      string
	JWTSecret string
	Now       func() time.Time
}
