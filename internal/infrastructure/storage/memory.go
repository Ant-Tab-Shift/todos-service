package storage

import (
	"context"
	"sync"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
)

type InMemory[V any] struct {
	rwm      sync.RWMutex
	serialID uint64
	data     map[uint64]V
}

func NewInMemory[V any]() *InMemory[V] {
	return &InMemory[V]{
		serialID: 1,
		data:     make(map[uint64]V),
	}
}

func (m *InMemory[V]) Save(ctx context.Context, value V, id uint64) (uint64, error) {
	err := ctx.Err()
	if err != nil {
		return 0, err
	}

	m.rwm.Lock()
	defer m.rwm.Unlock()

	if err = ctx.Err(); err != nil {
		return 0, err
	}

	if id == 0 {
		id = m.serialID
		m.serialID++
	}
	m.data[id] = value

	return id, nil
}

func (m *InMemory[V]) GetByID(ctx context.Context, id uint64) (V, error) {
	var (
		err  error
		zero V
	)
	if err = ctx.Err(); err != nil {
		return zero, err
	}

	m.rwm.RLock()
	defer m.rwm.RUnlock()

	if err = ctx.Err(); err != nil {
		return zero, err
	}

	value, ok := m.data[id]
	if !ok {
		return zero, domain.ErrNotExists
	}

	return value, nil
}

func (m *InMemory[V]) GetAll(ctx context.Context) ([]domain.Elem[V], error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	m.rwm.RLock()
	defer m.rwm.RUnlock()

	if err = ctx.Err(); err != nil {
		return nil, err
	}

	elems := make([]domain.Elem[V], 0, len(m.data))
	for id, value := range m.data {
		elem := domain.Elem[V]{
			ID: id,
			Value: value,
		}
		elems = append(elems, elem)
	}

	return elems, nil
}

func (m *InMemory[V]) Delete(ctx context.Context, id uint64) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	m.rwm.Lock()
	defer m.rwm.Unlock()

	if err = ctx.Err(); err != nil {
		return err
	}

	if _, ok := m.data[id]; !ok {
		return domain.ErrNotExists
	}
	delete(m.data, id)

	return nil
}
