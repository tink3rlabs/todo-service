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

// @openapi
// components:
//
//	schemas:
//	  PatchBody:
//	    type: object
//	    description: A JSONPatch document as defined by RFC 6902
//	    additionalProperties: false
//	    required:
//	      - op
//	      - path
//	    properties:
//	      op:
//	        type: string
//	        description: The operation to be performed
//	        enum:
//	          - add
//	          - remove
//	          - replace
//	          - move
//	          - copy
//	          - test
//	      path:
//	        type: string
//	        description: A JSON-Pointer
//	      value:
//	        description: The value to be used within the operations.
//	      from:
//	        type: string
//	        description: A string containing a JSON Pointer value.
type PatchBody struct{}
