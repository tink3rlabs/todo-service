package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"todo-service/features/todo"
	"todo-service/internal/middlewares"
	"todo-service/types"
)

type TodoRouter struct {
	Router    *chi.Mux
	service   *todo.TodoService
	formatter Formatter
	validator middlewares.Validator
}

// Define the JSON schemas as a map where the ctx(body, params and query) is the key and schema is the value
// Example: If you gave a request where you need to validate body, params and query
// var schema = map[string]string{
// 	"body": `{
// 		"type": "object",
// 		"properties": {
// 			"summary": { "type": "string" }
// 		},
// 		"required": ["summary"]
// 	}`,
// 	"params": `{
// 		"type": "object",
// 		"properties": {
// 			"id": { "type": "string" }
// 		},
// 		"required": ["id"]
// 	}`,
// 	"query": `{
// 		"type": "object",
// 		"properties": {
// 			"app": { "type": "string" }
// 		},
// 		"required": ["app"]
// 	}`,
// }

var createSchema = map[string]string{
	"body": `{
		"type": "object",
		"properties": {
			"summary": { "type": "string" },
			"done": { "type": "boolean" }
		},
		"required": ["summary"],
		"additionalProperties": false
	}`,
}

var putSchema = map[string]string{
	"body": `{
		"type": "object",
		"properties": {
			"summary": { "type": "string" },
			"done": { "type": "boolean" }
		},
		"required": ["summary", "done"],
		"additionalProperties": false
	}`,
	"params": `{
		"type": "object",
		"properties": {
			"id": { "type": "string" }
		},
		"required": ["id"]
	}`,
}

var idSchema = map[string]string{
	"params": `{
		"type": "object",
		"properties": {
			"id": { "type": "string" }
		},
		"required": ["id"]
	}`,
}

func NewTodoRouter() *TodoRouter {
	t := TodoRouter{}

	router := chi.NewRouter()
	router.Get("/{id}", t.validator.ValidateRequest(idSchema, t.GetTodo))
	router.Delete("/{id}", t.validator.ValidateRequest(idSchema, t.DeleteTodo))
	router.Put("/{id}", t.validator.ValidateRequest(putSchema, t.ReplaceTodo))
	router.Patch("/{id}", t.UpdateTodo)
	router.Post("/", t.validator.ValidateRequest(createSchema, t.CreateTodo))
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
//	    operationId: listTodos
//	    parameters:
//	      - name: limit
//	        in: query
//	        description: The number of todo items to return (defaults to 10)
//	        required: false
//	        schema:
//	          type: number
//	      - name: next
//	        in: query
//	        description: The next page identifier
//	        required: false
//	        schema:
//	          type: string
//	    responses:
//	      '200':
//	        description: successful operation
//	        content:
//	          application/json:
//	            schema:
//	              $ref: '#/components/schemas/TodoList'
func (t *TodoRouter) ListTodos(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("next")

	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if (err != nil) || limit <= 0 {
		limit = 10
	}

	todos, next, err := t.service.ListTodos(int(limit), cursor)
	t.formatter.Respond(types.TodoList{Todos: todos, Next: next}, err, w, r)
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
//	        description: The identifier of the Todo
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
//	        description: The identifier of the Todo
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
//	    operationId: createTodo
//	    requestBody:
//	      description: Create a new Todo
//	      content:
//	        application/json:
//	          schema:
//	            $ref: '#/components/schemas/TodoUpdate'
//	    responses:
//	      '201':
//	        description: successful operation
func (t *TodoRouter) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todoToCreate types.TodoUpdate
	decodeErr := json.NewDecoder(r.Body).Decode(&todoToCreate)
	if decodeErr != nil {
		t.formatter.Respond(nil, decodeErr, w, r)
	}

	todo, err := t.service.CreateTodo(todoToCreate)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	} else {
		render.Status(r, 201)
		render.JSON(w, r, todo)
	}
}

// @openapi
// paths:
//
//	/todos/{id}:
//	  put:
//	    tags:
//	      - todos
//	    summary: Replace a Todo
//	    description: Replace a Todo
//	    operationId: replaceTodo
//	    parameters:
//	      - name: id
//	        in: path
//	        description: The identifier of the Todo
//	        required: true
//	        schema:
//	          type: string
//	    requestBody:
//	      description: Updated Todo
//	      content:
//	        application/json:
//	          schema:
//	            $ref: '#/components/schemas/TodoUpdate'
//	    responses:
//	      '204':
//	        description: successful operation
func (t *TodoRouter) ReplaceTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var todoToUpdate types.TodoUpdate

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&todoToUpdate)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	todo := types.Todo{Id: id, Summary: todoToUpdate.Summary, Done: todoToUpdate.Done}
	err = t.service.UpdateTodo(todo)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	} else {
		render.NoContent(w, r)
	}
}

// @openapi
// paths:
//
//	/todos/{id}:
//	  patch:
//	    tags:
//	      - todos
//	    summary: Update a Todo
//	    description: Update a Todo
//	    operationId: updateTodo
//	    parameters:
//	      - name: id
//	        in: path
//	        description: The identifier of the Todo
//	        required: true
//	        schema:
//	          type: string
//	    requestBody:
//	      description: Updated Todo
//	      content:
//	        application/json-patch+json:
//	          schema:
//	            $ref: '#/components/schemas/PatchRequest'
//	    responses:
//	      '204':
//	        description: successful operation
func (t *TodoRouter) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	patch, err := jsonpatch.DecodePatch(body)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	currentRecord, err := t.service.GetTodo(id)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	currentBytes, err := json.Marshal(currentRecord)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	modifiedBytes, err := patch.Apply(currentBytes)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	var modified types.Todo
	err = json.Unmarshal(modifiedBytes, &modified)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	}

	if modified.Id != currentRecord.Id {
		t.formatter.Respond(nil, errors.New("bad request: id can't be changed"), w, r)
	}

	err = t.service.UpdateTodo(modified)
	if err != nil {
		t.formatter.Respond(nil, err, w, r)
	} else {
		render.NoContent(w, r)
	}
}
