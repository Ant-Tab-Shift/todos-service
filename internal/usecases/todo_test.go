package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
)

type mockTaskStorage struct {
	saveFunc    func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error)
	getByIDFunc func(ctx context.Context, id uint64) (domain.TaskSchema, error)
	getAllFunc  func(ctx context.Context) ([]domain.Elem[domain.TaskSchema], error)
	deleteFunc  func(ctx context.Context, id uint64) error
}

func (m *mockTaskStorage) Save(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, task, id)
	}
	return 0, nil
}

func (m *mockTaskStorage) GetByID(ctx context.Context, id uint64) (domain.TaskSchema, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return domain.TaskSchema{}, nil
}

func (m *mockTaskStorage) GetAll(ctx context.Context) ([]domain.Elem[domain.TaskSchema], error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockTaskStorage) Delete(ctx context.Context, id uint64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestNewTaskService(t *testing.T) {
	repo := &mockTaskStorage{}
	service := NewTaskService(repo)

	if service == nil {
		t.Fatal("NewTaskService returned nil")
	}
	if service.repo != repo {
		t.Error("repo not set correctly")
	}
}

func TestTaskService_Create_Success(t *testing.T) {
	expectedID := uint64(42)
	repo := &mockTaskStorage{
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			if id != 0 {
				t.Errorf("Save called with id = %d, want 0", id)
			}
			if task.Title != "Test Task" {
				t.Errorf("task.Title = %s, want 'Test Task'", task.Title)
			}
			if task.Description != "Test Description" {
				t.Errorf("task.Description = %s, want 'Test Description'", task.Description)
			}
			if task.IsDone != false {
				t.Error("task.IsDone should be false for new task")
			}
			return expectedID, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	task, err := service.Create(ctx, "Test Task", "Test Description")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if task.ID != expectedID {
		t.Errorf("task.ID = %d, want %d", task.ID, expectedID)
	}
	if task.Title != "Test Task" {
		t.Errorf("task.Title = %s, want 'Test Task'", task.Title)
	}
	if task.Description != "Test Description" {
		t.Errorf("task.Description = %s, want 'Test Description'", task.Description)
	}
	if task.IsDone != false {
		t.Error("task.IsDone should be false")
	}
}

func TestTaskService_Create_ValidationError(t *testing.T) {
	repo := &mockTaskStorage{
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			t.Error("Save should not be called when validation fails")
			return 0, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	_, err := service.Create(ctx, "", "Description")
	if err == nil {
		t.Error("Create should fail with empty title")
	}
}

func TestTaskService_Create_SaveError(t *testing.T) {
	expectedErr := errors.New("database error")
	repo := &mockTaskStorage{
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			return 0, expectedErr
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	_, err := service.Create(ctx, "Valid Title", "Valid Description")
	if err == nil {
		t.Fatal("Create should fail when Save fails")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestTaskService_GetByID_Success(t *testing.T) {
	expectedTask := domain.TaskSchema{
		Title:       "Test Task",
		Description: "Test Description",
		IsDone:      true,
	}
	expectedID := uint64(5)

	repo := &mockTaskStorage{
		getByIDFunc: func(ctx context.Context, id uint64) (domain.TaskSchema, error) {
			if id != expectedID {
				t.Errorf("GetByID called with id = %d, want %d", id, expectedID)
			}
			return expectedTask, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	task, err := service.GetByID(ctx, expectedID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if task.ID != expectedID {
		t.Errorf("task.ID = %d, want %d", task.ID, expectedID)
	}
	if task.Title != expectedTask.Title {
		t.Errorf("task.Title = %s, want %s", task.Title, expectedTask.Title)
	}
	if task.Description != expectedTask.Description {
		t.Errorf("task.Description = %s, want %s", task.Description, expectedTask.Description)
	}
	if task.IsDone != expectedTask.IsDone {
		t.Errorf("task.IsDone = %v, want %v", task.IsDone, expectedTask.IsDone)
	}
}

func TestTaskService_GetByID_NotFound(t *testing.T) {
	expectedErr := domain.ErrNotExists
	repo := &mockTaskStorage{
		getByIDFunc: func(ctx context.Context, id uint64) (domain.TaskSchema, error) {
			return domain.TaskSchema{}, expectedErr
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	_, err := service.GetByID(ctx, 999)
	if err == nil {
		t.Fatal("GetByID should fail for non-existent task")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestTaskService_GetAll_Success(t *testing.T) {
	expectedElems := []domain.Elem[domain.TaskSchema]{
		{
			ID: 1,
			Value: domain.TaskSchema{
				Title:       "Task 1",
				Description: "Description 1",
				IsDone:      false,
			},
		},
		{
			ID: 2,
			Value: domain.TaskSchema{
				Title:       "Task 2",
				Description: "Description 2",
				IsDone:      true,
			},
		},
		{
			ID: 3,
			Value: domain.TaskSchema{
				Title:       "Task 3",
				Description: "Description 3",
				IsDone:      false,
			},
		},
	}

	repo := &mockTaskStorage{
		getAllFunc: func(ctx context.Context) ([]domain.Elem[domain.TaskSchema], error) {
			return expectedElems, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	tasks, err := service.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(tasks) != len(expectedElems) {
		t.Fatalf("len(tasks) = %d, want %d", len(tasks), len(expectedElems))
	}

	for i, task := range tasks {
		expected := expectedElems[i]
		if task.ID != expected.ID {
			t.Errorf("tasks[%d].ID = %d, want %d", i, task.ID, expected.ID)
		}
		if task.Title != expected.Value.Title {
			t.Errorf("tasks[%d].Title = %s, want %s", i, task.Title, expected.Value.Title)
		}
		if task.Description != expected.Value.Description {
			t.Errorf("tasks[%d].Description = %s, want %s", i, task.Description, expected.Value.Description)
		}
		if task.IsDone != expected.Value.IsDone {
			t.Errorf("tasks[%d].IsDone = %v, want %v", i, task.IsDone, expected.Value.IsDone)
		}
	}
}

func TestTaskService_GetAll_Empty(t *testing.T) {
	repo := &mockTaskStorage{
		getAllFunc: func(ctx context.Context) ([]domain.Elem[domain.TaskSchema], error) {
			return []domain.Elem[domain.TaskSchema]{}, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	tasks, err := service.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("len(tasks) = %d, want 0", len(tasks))
	}
}

func TestTaskService_GetAll_Error(t *testing.T) {
	expectedErr := errors.New("database error")
	repo := &mockTaskStorage{
		getAllFunc: func(ctx context.Context) ([]domain.Elem[domain.TaskSchema], error) {
			return nil, expectedErr
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	_, err := service.GetAll(ctx)
	if err == nil {
		t.Fatal("GetAll should fail when repo fails")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestTaskService_Update_Success(t *testing.T) {
	taskID := uint64(10)
	originalTask := domain.TaskSchema{
		Title:       "Original Title",
		Description: "Original Description",
		IsDone:      false,
	}

	getByIDCalled := false
	saveCalled := false

	repo := &mockTaskStorage{
		getByIDFunc: func(ctx context.Context, id uint64) (domain.TaskSchema, error) {
			getByIDCalled = true
			if id != taskID {
				t.Errorf("GetByID called with id = %d, want %d", id, taskID)
			}
			return originalTask, nil
		},
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			saveCalled = true
			if id != taskID {
				t.Errorf("Save called with id = %d, want %d", id, taskID)
			}
			if task.Title != "Updated Title" {
				t.Errorf("task.Title = %s, want 'Updated Title'", task.Title)
			}
			if task.Description != "Updated Description" {
				t.Errorf("task.Description = %s, want 'Updated Description'", task.Description)
			}
			if task.IsDone != true {
				t.Error("task.IsDone should be true")
			}
			return id, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Update(ctx, taskID, "Updated Title", "Updated Description", true)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !getByIDCalled {
		t.Error("GetByID was not called")
	}
	if !saveCalled {
		t.Error("Save was not called")
	}
}

func TestTaskService_Update_NotFound(t *testing.T) {
	expectedErr := domain.ErrNotExists
	repo := &mockTaskStorage{
		getByIDFunc: func(ctx context.Context, id uint64) (domain.TaskSchema, error) {
			return domain.TaskSchema{}, expectedErr
		},
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			t.Error("Save should not be called when task not found")
			return 0, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Update(ctx, 999, "Title", "Description", false)
	if err == nil {
		t.Fatal("Update should fail for non-existent task")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestTaskService_Update_ValidationError(t *testing.T) {
	taskID := uint64(10)
	repo := &mockTaskStorage{
		getByIDFunc: func(ctx context.Context, id uint64) (domain.TaskSchema, error) {
			return domain.TaskSchema{
				Title:       "Original",
				Description: "Original",
				IsDone:      false,
			}, nil
		},
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			t.Error("Save should not be called when validation fails")
			return 0, nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Update(ctx, taskID, "", "Description", false)
	if err == nil {
		t.Error("Update should fail with invalid data")
	}
}

func TestTaskService_Update_SaveError(t *testing.T) {
	taskID := uint64(10)
	expectedErr := errors.New("save error")
	repo := &mockTaskStorage{
		getByIDFunc: func(ctx context.Context, id uint64) (domain.TaskSchema, error) {
			return domain.TaskSchema{
				Title:       "Original",
				Description: "Original",
				IsDone:      false,
			}, nil
		},
		saveFunc: func(ctx context.Context, task domain.TaskSchema, id uint64) (uint64, error) {
			return 0, expectedErr
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Update(ctx, taskID, "Valid Title", "Valid Description", true)
	if err == nil {
		t.Fatal("Update should fail when Save fails")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestTaskService_Delete_Success(t *testing.T) {
	taskID := uint64(15)
	deleteCalled := false

	repo := &mockTaskStorage{
		deleteFunc: func(ctx context.Context, id uint64) error {
			deleteCalled = true
			if id != taskID {
				t.Errorf("Delete called with id = %d, want %d", id, taskID)
			}
			return nil
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Delete(ctx, taskID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if !deleteCalled {
		t.Error("Delete was not called on repo")
	}
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	expectedErr := domain.ErrNotExists
	repo := &mockTaskStorage{
		deleteFunc: func(ctx context.Context, id uint64) error {
			return expectedErr
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Delete(ctx, 999)
	if err == nil {
		t.Fatal("Delete should fail for non-existent task")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestTaskService_Delete_Error(t *testing.T) {
	expectedErr := errors.New("delete error")
	repo := &mockTaskStorage{
		deleteFunc: func(ctx context.Context, id uint64) error {
			return expectedErr
		},
	}

	service := NewTaskService(repo)
	ctx := context.Background()

	err := service.Delete(ctx, 1)
	if err == nil {
		t.Fatal("Delete should fail when repo fails")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}
