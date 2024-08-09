package storage

import (
	"encoding/base64"
	"errors"
	"fmt"
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

func (m *MemoryAdapter) Execute(s string) error {
	return errors.New("memory adapter doesn't support executing arbitrary statements")
}

func (m *MemoryAdapter) Ping() error {
	return nil
}

func (m *MemoryAdapter) ListTodos(limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}
	nextId := ""

	id, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return todos, "", fmt.Errorf("failed to decode next cursor: %v", err)
	}

	// Get one extra item to be able to set that item's Id as the cursor for the next request
	for _, v := range m.todos {
		if (v.Id >= string(id)) && len(todos) < limit+1 {
			todos = append(todos, v)
		}
	}

	// If we have a full list, set the Id of the extra last item as the next cursor and remove it from the list of items to return
	if len(todos) == limit+1 {
		nextId = base64.StdEncoding.EncodeToString([]byte(todos[len(todos)-1].Id))
		todos = todos[:len(todos)-1]
	}

	return todos, nextId, nil
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
