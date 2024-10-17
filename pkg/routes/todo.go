package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"todo-service/pkg/features/todo"
	"todo-service/pkg/types"

	"github.com/tink3rlabs/magic/errors"
	"github.com/tink3rlabs/magic/middlewares"
)

type TodoRouter struct {
	Router  *chi.Mux
	service *todo.TodoService
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

var replaceSchema = map[string]string{
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
	h := middlewares.ErrorHandler{}
	v := middlewares.Validator{}

	router := chi.NewRouter()
	router.Get("/{id}", v.ValidateRequest(idSchema, h.Wrap(t.GetTodo)))
	router.Delete("/{id}", v.ValidateRequest(idSchema, h.Wrap(t.DeleteTodo)))
	router.Put("/{id}", v.ValidateRequest(replaceSchema, h.Wrap(t.ReplaceTodo)))
	router.Patch("/{id}", v.ValidateRequest(idSchema, h.Wrap(t.UpdateTodo)))
	router.Post("/", v.ValidateRequest(createSchema, h.Wrap(t.CreateTodo)))
	router.Get("/", h.Wrap(t.ListTodos))

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
//	      '500':
//	         $ref: '#/components/responses/ServerError'
func (t *TodoRouter) ListTodos(w http.ResponseWriter, r *http.Request) error {
	cursor := r.URL.Query().Get("next")

	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if (err != nil) || limit <= 0 {
		limit = 10
	}

	todos, next, err := t.service.ListTodos(int(limit), cursor)
	if err != nil {
		return err
	}
	render.JSON(w, r, types.TodoList{Todos: todos, Next: next})
	return nil
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
//	      '500':
//	         $ref: '#/components/responses/ServerError'
func (t *TodoRouter) GetTodo(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	todo, err := t.service.GetTodo(id)
	if err != nil {
		return err
	}
	render.JSON(w, r, todo)
	return nil
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
//	      '500':
//	         $ref: '#/components/responses/ServerError'
func (t *TodoRouter) DeleteTodo(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	err := t.service.DeleteTodo(id)
	if err != nil {
		return err
	}
	render.NoContent(w, r)
	return nil
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
//	      '400':
//	         $ref: '#/components/responses/BadRequest'
//	      '500':
//	         $ref: '#/components/responses/ServerError'
func (t *TodoRouter) CreateTodo(w http.ResponseWriter, r *http.Request) error {
	var todoToCreate types.TodoUpdate

	decodeErr := json.NewDecoder(r.Body).Decode(&todoToCreate)
	if decodeErr != nil {
		return decodeErr
	}

	todo, err := t.service.CreateTodo(todoToCreate)
	if err != nil {
		return err
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, todo)
	return nil
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
//	      '400':
//	         $ref: '#/components/responses/NotFound'
//	      '404':
//	         $ref: '#/components/responses/BadRequest'
//	      '500':
//	         $ref: '#/components/responses/ServerError'
func (t *TodoRouter) ReplaceTodo(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	var todoToUpdate types.TodoUpdate

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&todoToUpdate)
	if err != nil {
		return err
	}

	currentRecord, err := t.service.GetTodo(id)
	if err != nil {
		return &errors.NotFound{Message: "Todo not found"}
	}

	todo := types.Todo{Id: currentRecord.Id, Summary: todoToUpdate.Summary, Done: todoToUpdate.Done}
	err = t.service.UpdateTodo(todo)
	if err != nil {
		return err
	}

	render.NoContent(w, r)
	return nil
}

// @openapi
// paths:
//
//	/todos/{id}:
//	  patch:
//	    tags:
//	      - todos
//	    summary: Update a Todo
//	    description: Update a Todo using [JSON Patch](https://jsonpatch.com/)
//	    operationId: updateTodo
//	    parameters:
//	      - name: id
//	        in: path
//	        description: The identifier of the Todo
//	        required: true
//	        schema:
//	          type: string
//	    requestBody:
//	      description: JSON Patch operations to perform in order to update the Todo item
//	      content:
//	        application/json-patch+json:
//	          schema:
//	            type: array
//	            items:
//	              $ref: "#/components/schemas/PatchBody"
//	            example:
//	              - {"op": "replace", "path": "/summary", "value": "An updated TODO item summary"}
//	              - {"op": "replace", "path": "/done", "value": true}
//	    responses:
//	      '204':
//	        description: successful operation
//	      '400':
//	         $ref: '#/components/responses/NotFound'
//	      '404':
//	         $ref: '#/components/responses/BadRequest'
//	      '500':
//	         $ref: '#/components/responses/ServerError'
func (t *TodoRouter) UpdateTodo(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	patch, err := jsonpatch.DecodePatch(body)
	if err != nil {
		return &errors.BadRequest{Message: err.Error()}
	}

	currentRecord, err := t.service.GetTodo(id)
	if err != nil {
		return &errors.NotFound{Message: "Todo not found"}
	}

	currentBytes, err := json.Marshal(currentRecord)
	if err != nil {
		return err
	}

	modifiedBytes, err := patch.Apply(currentBytes)
	if err != nil {
		return &errors.BadRequest{Message: err.Error()}
	}

	var modified types.Todo
	err = json.Unmarshal(modifiedBytes, &modified)
	if err != nil {
		return err
	}

	if modified.Id != currentRecord.Id {
		return &errors.BadRequest{Message: "Id field can't be changed"}
	}

	err = t.service.UpdateTodo(modified)
	if err != nil {
		return err
	}

	render.NoContent(w, r)
	return nil
}
