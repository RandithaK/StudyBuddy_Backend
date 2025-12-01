package store

import (
	"errors"
	"sync"

	"github.com/RandithaK/StudyBuddy/backend/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

// In-memory thread-safe store
type InMemoryStore struct {
	mu      sync.RWMutex
	tasks   map[string]models.Task
	courses map[string]models.Course
	events  map[string]models.Event
	users   map[string]models.User
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		tasks:   make(map[string]models.Task),
		courses: make(map[string]models.Course),
		events:  make(map[string]models.Event),
		users:   make(map[string]models.User),
	}
}

// Task operations
func (s *InMemoryStore) GetTasks(userID string) []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if t.UserID == userID {
			res = append(res, t)
		}
	}
	return res
}

func (s *InMemoryStore) GetTask(id string) (models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.tasks[id]; ok {
		return t, nil
	}
	return models.Task{}, ErrNotFound
}

func (s *InMemoryStore) CreateTask(t models.Task) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
	return t
}

func (s *InMemoryStore) UpdateTask(id string, t models.Task) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return models.Task{}, ErrNotFound
	}
	t.ID = id
	s.tasks[id] = t
	return t, nil
}

func (s *InMemoryStore) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

// Course operations
func (s *InMemoryStore) GetCourses(userID string) []models.Course {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]models.Course, 0, len(s.courses))
	for _, c := range s.courses {
		if c.UserID == userID {
			// Calculate totalTasks and completedTasks for this course
			totalTasks := 0
			completedTasks := 0
			for _, t := range s.tasks {
				if t.UserID == userID && t.CourseID == c.ID {
					totalTasks++
					if t.Completed {
						completedTasks++
					}
				}
			}
			c.TotalTasks = totalTasks
			c.CompletedTasks = completedTasks
			res = append(res, c)
		}
	}
	return res
}

func (s *InMemoryStore) GetCourse(id string) (models.Course, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.courses[id]; ok {
		return c, nil
	}
	return models.Course{}, ErrNotFound
}

func (s *InMemoryStore) CreateCourse(c models.Course) models.Course {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.courses[c.ID] = c
	return c
}

// Event operations
func (s *InMemoryStore) GetEvents(userID string) []models.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]models.Event, 0, len(s.events))
	for _, e := range s.events {
		if e.UserID == userID {
			res = append(res, e)
		}
	}
	return res
}

func (s *InMemoryStore) CreateEvent(e models.Event) models.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[e.ID] = e
	return e
}

// User operations
func (s *InMemoryStore) GetUser(id string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return models.User{}, ErrNotFound
}

func (s *InMemoryStore) GetUserByEmail(email string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Email == email {
			return u, true
		}
	}
	return models.User{}, false
}

func (s *InMemoryStore) CreateUser(u models.User) models.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.ID] = u
	return u
}

func (s *InMemoryStore) GetUserByVerificationToken(token string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.VerificationToken == token {
			return u, nil
		}
	}
	return models.User{}, ErrNotFound
}

func (s *InMemoryStore) UpdateUser(id string, u models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}

	if u.Name != "" {
		existing.Name = u.Name
	}
	if u.Email != "" {
		existing.Email = u.Email
	}
	if u.IsVerified {
		existing.IsVerified = u.IsVerified
	}
	if u.VerificationToken != "" {
		existing.VerificationToken = u.VerificationToken
	}

	s.users[id] = existing
	return existing, nil
}

func (s *InMemoryStore) UpdateUserPassword(id string, hashedPassword string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}
	if hashedPassword == "" {
		return models.User{}, nil // nothing to update
	}
	existing.Password = hashedPassword
	s.users[id] = existing
	return existing, nil
}

func (s *InMemoryStore) MarkUserVerified(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.users[id]
	if !ok {
		return ErrNotFound
	}
	existing.IsVerified = true
	existing.VerificationToken = ""
	s.users[id] = existing
	return nil
}

// Notifications (Stub implementation for InMemoryStore)
func (s *InMemoryStore) GetNotifications(userID string) []models.Notification {
	return []models.Notification{}
}

func (s *InMemoryStore) GetNotificationByReferenceID(refID string, nType string) (models.Notification, error) {
	return models.Notification{}, ErrNotFound
}

func (s *InMemoryStore) CreateNotification(n models.Notification) models.Notification {
	return n
}

func (s *InMemoryStore) MarkNotificationAsRead(id string) error {
	return nil
}

func (s *InMemoryStore) GetUnreadNotificationsOlderThan(duration string) ([]models.Notification, error) {
	return []models.Notification{}, nil
}

func (s *InMemoryStore) MarkNotificationAsEmailed(id string) error {
	return nil
}

func (s *InMemoryStore) GetTasksDueIn(duration string) ([]models.Task, error) {
	return []models.Task{}, nil
}

func (s *InMemoryStore) GetEventsStartingIn(duration string) ([]models.Event, error) {
	return []models.Event{}, nil
}
