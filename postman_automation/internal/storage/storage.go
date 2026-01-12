package storage

import (
	"errors"
	"sync"

	"github.com/postman-automation/task-manager/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrTaskNotFound = errors.New("task not found")
	ErrUnauthorized = errors.New("unauthorized access to task")
)

type Storage struct {
	users        map[string]*model.User
	usersByEmail map[string]*model.User
	tasks        map[string]*model.Task
	mu           sync.RWMutex
}

func New() *Storage {
	return &Storage{
		users:        make(map[string]*model.User),
		usersByEmail: make(map[string]*model.User),
		tasks:        make(map[string]*model.Task),
	}
}

// User operations
func (s *Storage) CreateUser(user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[user.Email]; exists {
		return ErrUserExists
	}

	s.users[user.ID] = user
	s.usersByEmail[user.Email] = user
	return nil
}

func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *Storage) GetUserByID(id string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Task operations
func (s *Storage) CreateTask(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.ID] = task
	return nil
}

func (s *Storage) GetTask(taskID string) (*model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *Storage) GetUserTasks(userID string) []*model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*model.Task
	for _, task := range s.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (s *Storage) UpdateTask(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *Storage) DeleteTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; !exists {
		return ErrTaskNotFound
	}
	delete(s.tasks, taskID)
	return nil
}
