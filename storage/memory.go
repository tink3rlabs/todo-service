package storage

import (
	"sync"
	"todo-service/types"
)

var memoryAdapterLock = &sync.Mutex{}

type MemoryAdapter struct {
	todos []types.Todo
}

var memoryAdapterInstance *MemoryAdapter

func GetMemoryAdapterInstance() *MemoryAdapter {
	if memoryAdapterInstance == nil {
		memoryAdapterLock.Lock()
		defer memoryAdapterLock.Unlock()
		if memoryAdapterInstance == nil {
			memoryAdapterInstance = &MemoryAdapter{todos: []types.Todo{}}
		}
	}
	return memoryAdapterInstance
}

func (m *MemoryAdapter) ListTodos() ([]types.Todo, error) {
	return m.todos, nil
}

func (m *MemoryAdapter) GetTodo(id string) (types.Todo, error) {
	for _, v := range m.todos {
		if v.Id == id {
			return v, nil
		}
	}
	return types.Todo{}, ErrNotFound
}

func (m *MemoryAdapter) DeleteTodo(id string) error {
	for k, v := range m.todos {
		if v.Id == id {
			m.todos = append(m.todos[:k], m.todos[k+1:]...)
		}
	}
	return nil
}

func (m *MemoryAdapter) CreateTodo(todo types.Todo) error {
	m.todos = append(m.todos, todo)
	return nil
}
