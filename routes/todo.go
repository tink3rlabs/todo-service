package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"openapi-tester/features/todo"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func TodoRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/{id}", GetTodo)
	router.Delete("/{id}", DeleteTodo)
	router.Post("/", CreateTodo)
	router.Get("/", ListTodos)
	return router
}

var service = todo.New()

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
func ListTodos(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, service.ListTodos())
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
func GetTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	todo := service.GetTodo(id)
	if todo == nil {
		render.Status(r, 404)
		response := `{"status":"NOT_FOUND","error":"Todo not found"}`
		render.JSON(w, r, []byte(response))
	}
	render.JSON(w, r, service.GetTodo(id))
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
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	service.DeleteTodo(id)
	render.Status(r, 204)
	render.NoContent(w, r)
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
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todoToCreate todo.TodoUpdate
	err := json.NewDecoder(r.Body).Decode(&todoToCreate)
	if err != nil {
		errorMessage := fmt.Sprintf(`{"status":"BAD_REQUEST","error":"%v"}`, err)
		render.Status(r, 400)
		render.JSON(w, r, []byte(errorMessage))
	}
	todo := service.CreateTodo(todoToCreate)
	render.Status(r, 201)
	render.JSON(w, r, todo)
}
