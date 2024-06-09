package todo

import (
	"github.com/google/uuid"
)

// @openapi
// components:
//
//	schemas:
//	  Todo:
//	    type: object
//	    properties:
//	      id:
//	        type: string
//	        description: The Todo's identifier
//	        example: todo-1
//	      summary:
//	        type: string
//	        description: The Todo's summary
//	        example: Pick up the groceries
type Todo struct {
	Id      string `json:"id"`
	Summary string `json:"summary"`
}

// @openapi
// components:
//
//	schemas:
//	  TodoUpdate:
//	    type: object
//	    properties:
//	      summary:
//	        type: string
//	        description: The Todo's summary
//	        example: Pick up the groceries
type TodoUpdate struct {
	Summary string `json:"summary"`
}

type TodoService struct {
	todos []Todo
}

func New() TodoService {
	t := TodoService{todos: []Todo{}}
	return t
}

func (t *TodoService) ListTodos() []Todo {
	return t.todos
}

func (t *TodoService) GetTodo(id string) *Todo {
	for _, v := range t.todos {
		if v.Id == id {
			return &v
		}
	}
	return nil
}

func (t *TodoService) DeleteTodo(id string) {
	for k, v := range t.todos {
		if v.Id == id {
			t.todos = append(t.todos[:k], t.todos[k+1:]...)
		}
	}
}

func (t *TodoService) CreateTodo(todoToCreate TodoUpdate) *Todo {
	id := uuid.New()
	todo := Todo{
		Id:      id.String(),
		Summary: todoToCreate.Summary,
	}
	t.todos = append(t.todos, todo)
	return &todo
}
