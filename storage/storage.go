package storage

import (
	"errors"
	"todo-service/types"
)

var ErrNotFound = errors.New("not found")

type StorageAdapter interface {
	ListTodos() []types.Todo
	GetTodo(id string) (types.Todo, error)
	DeleteTodo(id string)
	CreateTodo(todo types.Todo)
}

type StorageAdapterType string
type StorageAdapterFactory struct{}

const (
	MEMORY StorageAdapterType = "memory"
	SQL    StorageAdapterType = "sql"
)

func (s StorageAdapterFactory) GetInstance(adapterType StorageAdapterType) (StorageAdapter, error) {
	switch adapterType {
	case MEMORY:
		return GetMemoryAdapterInstance(), nil
	case SQL:
		return GetSQLAdapterInstance(), nil
	default:
		return nil, errors.New("this storage adapter type isn't supported")
	}
}
