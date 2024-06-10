package types

// @openapi
// components:
//
//	responses:
//	  NotFound:
//	    description: The specified resource was not found
//	    content:
//	      application/json:
//	        schema:
//	          $ref: '#/components/schemas/Error'
//	  Unauthorized:
//	    description: Unauthorized
//	    content:
//	      application/json:
//	        schema:
//	          $ref: '#/components/schemas/Error'
//	schemas:
//	  Error:
//	    type: object
//	    properties:
//	      status:
//	        type: string
//	      error:
//	        type: string
type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
