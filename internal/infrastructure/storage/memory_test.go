package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
)

type testData struct {
	Name  string
	Value int
}

func TestNewInMemory(t *testing.T) {
	storage := NewInMemory[testData]()

	if storage == nil {
		t.Fatal("NewInMemory returned nil")
	}
	if storage.data == nil {
		t.Error("data map not initialized")
	}
	if storage.serialID != 1 {
		t.Errorf("serialID = %d, want 1", storage.serialID)
	}
}

func TestInMemory_Save_WithAutoID(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	data1 := testData{Name: "first", Value: 1}
	id1, err := storage.Save(ctx, data1, 0)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if id1 != 1 {
		t.Errorf("first id = %d, want 1", id1)
	}

	data2 := testData{Name: "second", Value: 2}
	id2, err := storage.Save(ctx, data2, 0)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if id2 != 2 {
		t.Errorf("second id = %d, want 2", id2)
	}

	if len(storage.data) != 2 {
		t.Errorf("data length = %d, want 2", len(storage.data))
	}
}

func TestInMemory_Save_WithSpecificID(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	data := testData{Name: "test", Value: 42}
	id, err := storage.Save(ctx, data, 100)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if id != 100 {
		t.Errorf("id = %d, want 100", id)
	}

	retrieved, err := storage.GetByID(ctx, 100)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != data.Name || retrieved.Value != data.Value {
		t.Errorf("retrieved = %+v, want %+v", retrieved, data)
	}
}

func TestInMemory_Save_Update(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	original := testData{Name: "original", Value: 1}
	id, err := storage.Save(ctx, original, 0)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	updated := testData{Name: "updated", Value: 2}
	_, err = storage.Save(ctx, updated, id)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, err := storage.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != "updated" || retrieved.Value != 2 {
		t.Errorf("retrieved = %+v, want updated data", retrieved)
	}
}

func TestInMemory_Save_CancelledContext(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data := testData{Name: "test", Value: 1}
	_, err := storage.Save(ctx, data, 0)
	if err == nil {
		t.Error("Save should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestInMemory_GetByID_Success(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	expected := testData{Name: "test", Value: 42}
	id, _ := storage.Save(ctx, expected, 0)

	result, err := storage.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if result != expected {
		t.Errorf("result = %+v, want %+v", result, expected)
	}
}

func TestInMemory_GetByID_NotExists(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	_, err := storage.GetByID(ctx, 999)
	if err == nil {
		t.Error("GetByID should fail for non-existent id")
	}
	if !errors.Is(err, domain.ErrNotExists) {
		t.Errorf("error = %v, want domain.ErrNotExists", err)
	}
}

func TestInMemory_GetByID_CancelledContext(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := storage.GetByID(ctx, 1)
	if err == nil {
		t.Error("GetByID should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestInMemory_GetAll_Empty(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	elems, err := storage.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(elems) != 0 {
		t.Errorf("len(elems) = %d, want 0", len(elems))
	}
}

func TestInMemory_GetAll_MultipleItems(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	data1 := testData{Name: "first", Value: 1}
	data2 := testData{Name: "second", Value: 2}
	data3 := testData{Name: "third", Value: 3}

	id1, _ := storage.Save(ctx, data1, 0)
	id2, _ := storage.Save(ctx, data2, 0)
	id3, _ := storage.Save(ctx, data3, 0)

	elems, err := storage.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(elems) != 3 {
		t.Errorf("len(elems) = %d, want 3", len(elems))
	}

	found := make(map[uint64]bool)
	for _, elem := range elems {
		found[elem.ID] = true
		switch elem.ID {
		case id1:
			if elem.Value != data1 {
				t.Errorf("elem[%d] = %+v, want %+v", id1, elem.Value, data1)
			}
		case id2:
			if elem.Value != data2 {
				t.Errorf("elem[%d] = %+v, want %+v", id2, elem.Value, data2)
			}
		case id3:
			if elem.Value != data3 {
				t.Errorf("elem[%d] = %+v, want %+v", id3, elem.Value, data3)
			}
		default:
			t.Errorf("unexpected id: %d", elem.ID)
		}
	}

	if !found[id1] || !found[id2] || !found[id3] {
		t.Error("not all elements found in GetAll result")
	}
}

func TestInMemory_GetAll_CancelledContext(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := storage.GetAll(ctx)
	if err == nil {
		t.Error("GetAll should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestInMemory_Delete_Success(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	data := testData{Name: "test", Value: 1}
	id, _ := storage.Save(ctx, data, 0)

	err := storage.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.GetByID(ctx, id)
	if !errors.Is(err, domain.ErrNotExists) {
		t.Error("item should not exist after deletion")
	}

	if len(storage.data) != 0 {
		t.Errorf("data length = %d, want 0", len(storage.data))
	}
}

func TestInMemory_Delete_NotExists(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	err := storage.Delete(ctx, 999)
	if err == nil {
		t.Error("Delete should fail for non-existent id")
	}
	if !errors.Is(err, domain.ErrNotExists) {
		t.Errorf("error = %v, want domain.ErrNotExists", err)
	}
}

func TestInMemory_Delete_CancelledContext(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := storage.Delete(ctx, 1)
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestInMemory_ConcurrentAccess(t *testing.T) {
	storage := NewInMemory[testData]()
	ctx := context.Background()

	done := make(chan bool)
	operations := 100

	for i := 0; i < operations; i++ {
		go func(val int) {
			data := testData{Name: "concurrent", Value: val}
			_, err := storage.Save(ctx, data, 0)
			if err != nil {
				t.Errorf("concurrent Save failed: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < operations; i++ {
		go func() {
			_, _ = storage.GetAll(ctx)
			done <- true
		}()
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < operations*2; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("concurrent operations timed out")
		}
	}

	elems, err := storage.GetAll(ctx)
	if err != nil {
		t.Fatalf("final GetAll failed: %v", err)
	}
	if len(elems) != operations {
		t.Errorf("final count = %d, want %d", len(elems), operations)
	}
}