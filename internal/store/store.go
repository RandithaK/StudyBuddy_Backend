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
