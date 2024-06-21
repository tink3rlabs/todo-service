package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"todo-service/features/todo"
	"todo-service/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type TodoRouter struct {
	Router    *chi.Mux
	service   *todo.TodoService
	formatter Formatter
}

func NewTodoRouter() *TodoRouter {
	t := TodoRouter{}

	router := chi.NewRouter()
	router.Get("/{id}", t.GetTodo)
	router.Delete("/{id}", t.DeleteTodo)
	router.Post("/", t.CreateTodo)
	router.Get("/", t.ListTodos)
	t.Router = router
	t.service = todo.NewTodoService()

	return &t
}

// @openapi
// paths:
//
//	/todos:
//	  get:
//	    tags:
//	      - todos
//	    summary: Get all Todos
//	    description: Returns all Todos
//	    operationId: getTodos
//	    responses:
//	      '200':
//	        description: successful operation
//	        content:
//	          application/json:
//	            schema:
//	              type: array
//	              items:
//	                $ref: '#/components/schemas/Todo'
func (t *TodoRouter) ListTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := t.service.ListTodos()
	t.formatter.Respond(todos, err, w, r)
}

// @openapi
// paths:
//
//	/todos/{id}:
//	  get:
//	    tags:
//	      - todos
//	    summary: Get a single Todo
//	    description: Returns a Todos with the identifier {id} if exists
//	    operationId: getTodo
//	    parameters:
//	      - name: id
//	        in: path
//	        description: The identifier of the Todo to retrieve
//	        required: true
//	        schema:
//	          type: string
//	    responses:
//	      '200':
//	        description: successful operation
//	        content:
//	          application/json:
//	            schema:
//	              $ref: '#/components/schemas/Todo'
//	      '404':
//	         $ref: '#/components/responses/NotFound'
func (t *TodoRouter) GetTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	todo, err := t.service.GetTodo(id)
	t.formatter.Respond(todo, err, w, r)
}

// @openapi
// paths:
//
//	/todos/{id}:
//	  delete:
//	    tags:
//	      - todos
//	    summary: Delete a single Todo
//	    description: Deletes a Todos with the identifier {id} if exists
//	    operationId: deleteTodo
//	    parameters:
//	      - name: id
//	        in: path
//	        description: The identifier of the Todo to delete
//	        required: true
//	        schema:
//	          type: string
//	    responses:
//	      '204':
//	        description: successful operation
func (t *TodoRouter) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := t.service.DeleteTodo(id)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	} else {
		render.Status(r, 204)
		render.NoContent(w, r)
	}
}

// @openapi
// paths:
//
//	/todos:
//	  post:
//	    tags:
//	      - todos
//	    summary: Create a Todo
//	    description: Create a new Todo
//	    operationId: addTodo
//	    requestBody:
//	      description: Create a new pet in the store
//	      content:
//	        application/json:
//	          schema:
//	            $ref: '#/components/schemas/TodoUpdate'
//	    responses:
//	      '201':
//	        description: successful operation
func (t *TodoRouter) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todoToCreate types.TodoUpdate
	err := json.NewDecoder(r.Body).Decode(&todoToCreate)
	if err != nil {
		errorMessage := fmt.Sprintf(`{"status":"BAD_REQUEST","error":"%v"}`, err)
		render.Status(r, 400)
		render.JSON(w, r, []byte(errorMessage))
	}
	todo, err := t.service.CreateTodo(todoToCreate)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	} else {
		render.Status(r, 201)
		render.JSON(w, r, todo)
	}
}
