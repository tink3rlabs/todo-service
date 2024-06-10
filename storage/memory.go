package storage

import (
	"fmt"
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
			fmt.Println("Creating single instance now.")
			instance = &MemoryAdapter{}
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}

	return instance
}

func (m *MemoryAdapter) ListTodos() []types.Todo {
	return m.todos
}

func (m *MemoryAdapter) GetTodo(id string) *types.Todo {
	for _, v := range m.todos {
		if v.Id == id {
			return &v
		}
	}
	return nil
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
