package todo

import (
	"todo-service/storage"
	"todo-service/types"

	"github.com/google/uuid"
)

type TodoService struct {
	storage storage.StorageAdapter
}

func NewTodoService() *TodoService {
	s := storage.StorageAdapterFactory{}
	storageAdapter, err := s.GetInstance(storage.MEMORY)
	if err != nil {
		return nil
	}
	t := TodoService{storage: storageAdapter}
	return &t
}

func (t *TodoService) ListTodos() []types.Todo {
	return t.storage.ListTodos()
}

func (t *TodoService) GetTodo(id string) (types.Todo, error) {
	todo, err := t.storage.GetTodo(id)
	if err != nil {
		return todo, err
	}
	return todo, nil
}

func (t *TodoService) DeleteTodo(id string) {
	t.storage.DeleteTodo(id)
}

func (t *TodoService) CreateTodo(todoToCreate types.TodoUpdate) types.Todo {
	id := uuid.New()
	todo := types.Todo{
		Id:      id.String(),
		Summary: todoToCreate.Summary,
	}
	t.storage.CreateTodo(todo)
	return todo
}
