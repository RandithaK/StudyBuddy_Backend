package main

import "context"

// Store defines the repository interface used by handlers.
type Store interface {
	// Tasks
	GetTasks() []Task
	GetTask(id string) (Task, error)
	CreateTask(t Task) Task
	UpdateTask(id string, t Task) (Task, error)
	DeleteTask(id string) error

	// Courses
	GetCourses() []Course
	CreateCourse(c Course) Course

	// Events
	GetEvents() []Event
	CreateEvent(e Event) Event

	// Users
	GetUser(id string) (User, error)
	GetUserByEmail(email string) (User, bool)
	CreateUser(u User) User
}

// NewStore returns a Store implementation. If MONGO_URI is provided, a MongoStore will be used.
func NewStore(ctx context.Context, uri string) (Store, error) {
	if uri != "" {
		ms, err := NewMongoStore(ctx, uri, "studybuddy")
		if err != nil {
			return nil, err
		}
		return ms, nil
	}
	return NewInMemoryStore(), nil
}
