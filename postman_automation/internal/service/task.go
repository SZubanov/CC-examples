package service

import (
	"errors"
	"time"

	"github.com/postman-automation/task-manager/internal/model"
	"github.com/postman-automation/task-manager/internal/storage"
)

var (
	ErrInvalidStatus   = errors.New("invalid status")
	ErrInvalidPriority = errors.New("invalid priority")
)

type TaskService struct {
	storage *storage.Storage
}

func NewTaskService(storage *storage.Storage) *TaskService {
	return &TaskService{storage: storage}
}

func (s *TaskService) CreateTask(userID string, req model.CreateTaskRequest) (*model.TaskResponse, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	if !isValidPriority(req.Priority) {
		return nil, ErrInvalidPriority
	}

	task := model.NewTask(userID, req.Title, req.Description, req.Priority)
	if err := s.storage.CreateTask(task); err != nil {
		return nil, err
	}

	resp := task.ToResponse()
	return &resp, nil
}

func (s *TaskService) GetTask(taskID, userID string) (*model.TaskResponse, error) {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	if task.UserID != userID {
		return nil, storage.ErrUnauthorized
	}

	resp := task.ToResponse()
	return &resp, nil
}

func (s *TaskService) GetUserTasks(userID string) []model.TaskResponse {
	tasks := s.storage.GetUserTasks(userID)
	responses := make([]model.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, task.ToResponse())
	}
	return responses
}

func (s *TaskService) UpdateTask(taskID, userID string, req model.UpdateTaskRequest) (*model.TaskResponse, error) {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	if task.UserID != userID {
		return nil, storage.ErrUnauthorized
	}

	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	if !isValidPriority(req.Priority) {
		return nil, ErrInvalidPriority
	}

	task.Title = req.Title
	task.Description = req.Description
	task.Priority = req.Priority
	task.UpdatedAt = time.Now()

	if err := s.storage.UpdateTask(task); err != nil {
		return nil, err
	}

	resp := task.ToResponse()
	return &resp, nil
}

func (s *TaskService) UpdateStatus(taskID, userID string, req model.UpdateStatusRequest) (*model.TaskResponse, error) {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	if task.UserID != userID {
		return nil, storage.ErrUnauthorized
	}

	if !isValidStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	task.Status = req.Status
	task.UpdatedAt = time.Now()

	if err := s.storage.UpdateTask(task); err != nil {
		return nil, err
	}

	resp := task.ToResponse()
	return &resp, nil
}

func (s *TaskService) DeleteTask(taskID, userID string) error {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return err
	}

	if task.UserID != userID {
		return storage.ErrUnauthorized
	}

	return s.storage.DeleteTask(taskID)
}

func isValidStatus(status model.TaskStatus) bool {
	return status == model.StatusPending ||
		status == model.StatusInProgress ||
		status == model.StatusDone
}

func isValidPriority(priority model.TaskPriority) bool {
	return priority == model.PriorityLow ||
		priority == model.PriorityMedium ||
		priority == model.PriorityHigh
}
