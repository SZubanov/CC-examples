package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/postman-automation/task-manager/internal/middleware"
	"github.com/postman-automation/task-manager/internal/model"
	"github.com/postman-automation/task-manager/internal/service"
	"github.com/postman-automation/task-manager/internal/storage"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Title is required")
		return
	}

	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	}

	resp, err := h.taskService.CreateTask(userID, req)
	if err != nil {
		if err == service.ErrInvalidPriority {
			respondWithError(w, http.StatusBadRequest, "validation_error", "Priority must be low, medium, or high")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	vars := mux.Vars(r)
	taskID := vars["id"]

	resp, err := h.taskService.GetTask(taskID, userID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, "task_not_found", "Task not found")
			return
		}
		if err == storage.ErrUnauthorized {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", "You don't have access to this task")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	tasks := h.taskService.GetUserTasks(userID)
	respondWithJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	vars := mux.Vars(r)
	taskID := vars["id"]

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Title is required")
		return
	}

	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	}

	resp, err := h.taskService.UpdateTask(taskID, userID, req)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, "task_not_found", "Task not found")
			return
		}
		if err == storage.ErrUnauthorized {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", "You don't have access to this task")
			return
		}
		if err == service.ErrInvalidPriority {
			respondWithError(w, http.StatusBadRequest, "validation_error", "Priority must be low, medium, or high")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	vars := mux.Vars(r)
	taskID := vars["id"]

	var req model.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Status == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Status is required")
		return
	}

	resp, err := h.taskService.UpdateStatus(taskID, userID, req)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, "task_not_found", "Task not found")
			return
		}
		if err == storage.ErrUnauthorized {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", "You don't have access to this task")
			return
		}
		if err == service.ErrInvalidStatus {
			respondWithError(w, http.StatusBadRequest, "validation_error", "Status must be pending, in_progress, or done")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	vars := mux.Vars(r)
	taskID := vars["id"]

	err := h.taskService.DeleteTask(taskID, userID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, "task_not_found", "Task not found")
			return
		}
		if err == storage.ErrUnauthorized {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", "You don't have access to this task")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getUserID(r *http.Request) string {
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	return userID
}
