package store

import (
	"context"

	"github.com/RandithaK/StudyBuddy_Backend/pkg/models"
)

// Store defines the repository interface used by handlers.
type Store interface {
	// Tasks
	GetTasks(userID string) []models.Task
	GetTask(id string) (models.Task, error)
	CreateTask(t models.Task) models.Task
	UpdateTask(id string, t models.Task) (models.Task, error)
	DeleteTask(id string) error

	// Courses
	GetCourses(userID string) []models.Course
	GetCourse(id string) (models.Course, error)
	CreateCourse(c models.Course) models.Course

	// Events
	GetEvents(userID string) []models.Event
	CreateEvent(e models.Event) models.Event

	// Users
	GetUser(id string) (models.User, error)
	GetUserByEmail(email string) (models.User, bool)
	GetUserByVerificationToken(token string) (models.User, error)
	CreateUser(u models.User) models.User
	UpdateUser(id string, u models.User) (models.User, error)
	UpdateUserPassword(id string, hashedPassword string) (models.User, error)
	MarkUserVerified(id string) error

	// Notifications
	GetNotifications(userID string) []models.Notification
	GetNotificationByReferenceID(refID string, nType string) (models.Notification, error)
	CreateNotification(n models.Notification) models.Notification
	MarkNotificationAsRead(id string) error
	GetUnreadNotificationsOlderThan(duration string) ([]models.Notification, error)
	GetUnreadNotificationsOlderThanForUser(userID string, duration string) ([]models.Notification, error)
	MarkNotificationAsEmailed(id string) error

	// Worker Helpers
	GetTasksDueIn(duration string) ([]models.Task, error)
	GetEventsStartingIn(duration string) ([]models.Event, error)
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
