package todo

import (
	"log/slog"

	"github.com/google/uuid"

	"todo-service/internal/logger"
	"todo-service/internal/storage"
	"todo-service/types"
)

type TodoService struct {
	storage storage.StorageAdapter
}

func NewTodoService() *TodoService {
	s := storage.StorageAdapterFactory{}
	storageAdapter, err := s.GetInstance(storage.DEFAULT)
	if err != nil {
		logger.Fatal("failed to create TodoService instance", slog.Any("error", err.Error()))
	}
	t := TodoService{storage: storageAdapter}
	return &t
}

func (t *TodoService) ListTodos(limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}
	next, err := t.storage.List(&todos, "Id", limit, cursor)

	return todos, next, err
}

func (t *TodoService) GetTodo(id string) (types.Todo, error) {
	todo := types.Todo{}
	err := t.storage.Get(&todo, "Id", id)
	return todo, err
}

func (t *TodoService) DeleteTodo(id string) error {
	return t.storage.Delete(&types.Todo{}, "Id", id)
}

func (t *TodoService) UpdateTodo(todoToUpdate types.Todo) error {
	return t.storage.Update(todoToUpdate, "Id", todoToUpdate.Id)
}

func (t *TodoService) CreateTodo(todoToCreate types.TodoUpdate) (types.Todo, error) {
	todo := types.Todo{}

	// Using UUIDv7 in order to easily support cursor based pagination without extra fields
	//
	// From the RFC (https://datatracker.ietf.org/doc/rfc9562/)
	//
	// UUIDv7 features a time-ordered value field derived from the widely
	// implemented and well-known Unix Epoch timestamp source, the number of
	// milliseconds since midnight 1 Jan 1970 UTC, leap seconds excluded.
	// Generally, UUIDv7 has improved entropy characteristics over UUIDv1
	// (Section 5.1) or UUIDv6 (Section 5.6).
	//
	// UUIDv7 values are created by allocating a Unix timestamp in
	// milliseconds in the most significant 48 bits and filling the
	// remaining 74 bits, excluding the required version and variant bits,
	// with random bits for each new UUIDv7 generated to provide uniqueness
	// as per Section 6.9.
	id, err := uuid.NewV7()
	if err != nil {
		return todo, err
	}

	todo.Id = id.String()
	todo.Summary = todoToCreate.Summary
	todo.Done = todoToCreate.Done

	err = t.storage.Create(todo)
	return todo, err
}
