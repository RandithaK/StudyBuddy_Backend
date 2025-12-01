package main

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

// In-memory thread-safe store
type InMemoryStore struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	courses map[string]Course
	events  map[string]Event
	users   map[string]User
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		tasks:   make(map[string]Task),
		courses: make(map[string]Course),
		events:  make(map[string]Event),
		users:   make(map[string]User),
	}
}

// Task operations
func (s *InMemoryStore) GetTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		res = append(res, t)
	}
	return res
}

func (s *InMemoryStore) GetTask(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.tasks[id]; ok {
		return t, nil
	}
	return Task{}, ErrNotFound
}

func (s *InMemoryStore) CreateTask(t Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
	return t
}

func (s *InMemoryStore) UpdateTask(id string, t Task) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return Task{}, ErrNotFound
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
func (s *InMemoryStore) GetCourses() []Course {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Course, 0, len(s.courses))
	for _, c := range s.courses {
		res = append(res, c)
	}
	return res
}

func (s *InMemoryStore) CreateCourse(c Course) Course {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.courses[c.ID] = c
	return c
}

// Event operations
func (s *InMemoryStore) GetEvents() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Event, 0, len(s.events))
	for _, e := range s.events {
		res = append(res, e)
	}
	return res
}

func (s *InMemoryStore) CreateEvent(e Event) Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[e.ID] = e
	return e
}

// User operations
func (s *InMemoryStore) GetUser(id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

func (s *InMemoryStore) GetUserByEmail(email string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Email == email {
			return u, true
		}
	}
	return User{}, false
}

func (s *InMemoryStore) CreateUser(u User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.ID] = u
	return u
}
