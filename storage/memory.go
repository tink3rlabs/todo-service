package storage

import (
	"sync"
	"todo-service/types"
)

var lock = &sync.Mutex{}

type MemoryAdapter struct {
	todos []types.Todo
}

var instance *MemoryAdapter

func GetMemoryAdapterInstance() *MemoryAdapter {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = &MemoryAdapter{todos: []types.Todo{}}
		}
	}
	return instance
}

func (m *MemoryAdapter) ListTodos() []types.Todo {
	return m.todos
}

func (m *MemoryAdapter) GetTodo(id string) (types.Todo, error) {
	for _, v := range m.todos {
		if v.Id == id {
			return v, nil
		}
	}
	return types.Todo{}, ErrNotFound
}

func (m *MemoryAdapter) DeleteTodo(id string) {
	for k, v := range m.todos {
		if v.Id == id {
			m.todos = append(m.todos[:k], m.todos[k+1:]...)
		}
	}
}

func (m *MemoryAdapter) CreateTodo(todo types.Todo) {
	m.todos = append(m.todos, todo)
}
