package usecases

import (
	"context"
	"fmt"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
	"github.com/Ant-Tab-Shift/todos-service/internal/usecases/utils"
)

type TaskStorage interface {
	Save(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error)
	GetByID(ctx context.Context, id uint64) (domain.TaskSchema, error)
	GetAll(ctx context.Context) ([]domain.Elem[domain.TaskSchema], error)
	Delete(ctx context.Context, id uint64) error
}

type TaskService struct {
	repo     TaskStorage
}

func NewTaskService(repo TaskStorage) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, title, description string) (domain.Task, error) {
	task := domain.TaskSchema{
		Title: title,
		Description: description,
		IsDone: false,
	}
	err := utils.Validate(&task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("validation failed: %w", err)
	}

	id, err := s.repo.Save(ctx, task, 0)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to save task: %w", err)
	}

	return domain.Task{
		ID: id,
		TaskSchema: task,
	}, nil
}

func (s *TaskService) GetByID(ctx context.Context, id uint64) (domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	return domain.Task{
		ID: id,
		TaskSchema: task,
	}, nil
}

func (s *TaskService) GetAll(ctx context.Context) ([]domain.Task, error) {
	elems, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	tasks := make([]domain.Task, len(elems))
	for i, elem := range elems {
		tasks[i] = domain.Task{
			ID: elem.ID,
			TaskSchema: elem.Value,
		}
	}

	return tasks, nil
}

func (s *TaskService) Update(ctx context.Context, id uint64, title, description string, isDone bool) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	task.Title = title
	task.Description = description
	task.IsDone = isDone

	if err := utils.Validate(&task); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if _, err := s.repo.Save(ctx, task, id); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *TaskService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}
