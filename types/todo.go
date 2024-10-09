package types

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
//	        example: 01909c42-cc90-75dc-a943-2d87a16e787d
//	      summary:
//	        type: string
//	        description: The Todo's summary
//	        example: Pick up the groceries
//	      done:
//	        type: boolean
//	        description: An indicator that tells if the Todo item is complete
//	        example: false
type Todo struct {
	Id      string `json:"id"`
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
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
//	      done:
//	        type: boolean
//	        description: An indicator that tells if the Todo item is complete
//	        example: false
type TodoUpdate struct {
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
}

// @openapi
// components:
//
//	schemas:
//	  TodoList:
//	    type: object
//	    properties:
//	      todos:
//	        type: array
//	        items:
//	          $ref: '#/components/schemas/Todo'
//	      next:
//	        type: string
//	        description: An identifier to use when requesting the next set of todos
//	        example: MDE5MDlhOGUtNjcwNi03NWY1LWJjMjUtNWM0MjY0ZjUwZTQ1
type TodoList struct {
	Todos []Todo `json:"todos"`
	Next  string `json:"next"`
}
