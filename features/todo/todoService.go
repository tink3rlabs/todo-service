package todo

import (
	"log"
	"todo-service/storage"
	"todo-service/types"

	"github.com/google/uuid"
)

type TodoService struct {
	storage storage.StorageAdapter
}

func NewTodoService() *TodoService {
	s := storage.StorageAdapterFactory{}
	storageAdapter, err := s.GetInstance(storage.DEFAULT)
	if err != nil {
		log.Fatalf("failed to create TodoService instance: %s", err.Error())
		return nil
	}
	t := TodoService{storage: storageAdapter}
	return &t
}

func (t *TodoService) ListTodos() ([]types.Todo, error) {
	return t.storage.ListTodos()
}

func (t *TodoService) GetTodo(id string) (types.Todo, error) {
	return t.storage.GetTodo(id)
}

func (t *TodoService) DeleteTodo(id string) error {
	return t.storage.DeleteTodo(id)
}

func (t *TodoService) CreateTodo(todoToCreate types.TodoUpdate) (types.Todo, error) {
	id := uuid.New()
	todo := types.Todo{
		Id:      id.String(),
		Summary: todoToCreate.Summary,
	}
	err := t.storage.CreateTodo(todo)
	return todo, err
}
