package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
	"github.com/Ant-Tab-Shift/todos-service/internal/transport/http/dto"
	"github.com/Ant-Tab-Shift/todos-service/internal/transport/http/models"
)

type TaskService interface {
	Create(ctx context.Context, title, description string) (domain.Task, error)
	GetByID(ctx context.Context, id uint64) (domain.Task, error)
	GetAll(ctx context.Context) ([]domain.Task, error)
	Update(ctx context.Context, id uint64, title, description string, completed bool) error
	Delete(ctx context.Context, id uint64) error
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, dto.ErrorResponse{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	task, err := h.service.Create(r.Context(), req.Title, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyTitle) {
			writeJSON(w, dto.ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		writeJSON(w, dto.ErrorResponse{Error: "internal server error"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, dto.ToTaskResponse(&task), http.StatusCreated)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		writeJSON(w, dto.ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotExists) {
			writeJSON(w, dto.ErrorResponse{Error: "task not found"}, http.StatusNotFound)
			return
		}
		writeJSON(w, dto.ErrorResponse{Error: "internal server error"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, dto.ToTaskResponse(&task), http.StatusOK)
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetAll(r.Context())
	if err != nil {
		writeJSON(w, dto.ErrorResponse{Error: "internal server error"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, dto.ToTaskListResponse(tasks), http.StatusOK)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		writeJSON(w, dto.ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	var req dto.UpdateTaskRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, dto.ErrorResponse{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	if err = h.service.Update(r.Context(), id, req.Title, req.Description, req.IsDone); err != nil {
		if errors.Is(err, domain.ErrNotExists) {
			writeJSON(w, dto.ErrorResponse{Error: "task not found"}, http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrEmptyTitle) {
			writeJSON(w, dto.ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		writeJSON(w, dto.ErrorResponse{Error: "internal server error"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r)
	if err != nil {
		writeJSON(w, dto.ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	if err = h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotExists) {
			writeJSON(w, dto.ErrorResponse{Error: "task not found"}, http.StatusNotFound)
			return
		}
		writeJSON(w, dto.ErrorResponse{Error: "internal server error"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) Handlers() []models.Endpoint {
	return []models.Endpoint{
		{Pattern: "POST /todos", Func: h.Create},
		{Pattern: "GET /todos/{id}", Func: h.GetByID},
		{Pattern: "GET /todos", Func: h.GetAll},
		{Pattern: "PUT /todos/{id}", Func: h.Update},
		{Pattern: "DELETE /todos/{id}", Func: h.Delete},
	}
}

func parseIDFromPath(r *http.Request) (uint64, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, errors.New("task id is required")
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id format")
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// логирование ошибки (TODO: использовать logger)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
