package storage

import (
	"embed"
	"errors"

	"todo-service/types"

	"github.com/spf13/viper"
)

var ConfigFs embed.FS
var ErrNotFound = errors.New("not found")

type StorageAdapter interface {
	Execute(statement string) error
	ListTodos() ([]types.Todo, error)
	GetTodo(id string) (types.Todo, error)
	DeleteTodo(id string) error
	CreateTodo(todo types.Todo) error
}

type StorageAdapterType string
type StorageProviders string
type StorageAdapterFactory struct{}

const (
	DEFAULT StorageAdapterType = "default"
	MEMORY  StorageAdapterType = "memory"
	SQL     StorageAdapterType = "sql"
)

const (
	POSTGRESQL StorageProviders = "postgresql"
	MYSQL      StorageProviders = "mysql"
	SQLITE     StorageProviders = "sqlite"
)

func (s StorageAdapterFactory) GetInstance(adapterType StorageAdapterType) (StorageAdapter, error) {
	if adapterType == DEFAULT {
		adapterType = StorageAdapterType(viper.GetString("storage.type"))
	}
	switch adapterType {
	case MEMORY:
		return GetMemoryAdapterInstance(), nil
	case SQL:
		return GetSQLAdapterInstance(), nil
	default:
		return nil, errors.New("this storage adapter type isn't supported")
	}
}
