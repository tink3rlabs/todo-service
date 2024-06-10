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
