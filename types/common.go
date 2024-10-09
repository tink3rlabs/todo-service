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
//	  PatchRequest:
//	    type: array
//	    items:
//	      oneOf:
//	        - $ref: '#/components/schemas/JSONPatchRequestAddReplaceTest'
//	        - $ref: '#/components/schemas/JSONPatchRequestRemove'
//	        - $ref: '#/components/schemas/JSONPatchRequestMoveCopy'
//	  JSONPatchRequestAddReplaceTest:
//	    type: object
//	    additionalProperties: false
//	    required:
//	      - value
//	      - op
//	      - path
//	    properties:
//	      path:
//	        description: A JSON Pointer path.
//	        type: string
//	      value:
//	        description: The value to add, replace or test.
//	      op:
//	        description: The operation to perform.
//	        type: string
//	        enum:
//	          - add
//	          - replace
//	          - test
//	  JSONPatchRequestRemove:
//	    type: object
//	    additionalProperties: false
//	    required:
//	      - op
//	      - path
//	    properties:
//	      path:
//	        description: A JSON Pointer path.
//	        type: string
//	      op:
//	        description: The operation to perform.
//	        type: string
//	        enum:
//	          - remove
//	  JSONPatchRequestMoveCopy:
//	    type: object
//	    additionalProperties: false
//	    required:
//	      - from
//	      - op
//	      - path
//	    properties:
//	      path:
//	        description: A JSON Pointer path.
//	        type: string
//	      op:
//	        description: The operation to perform.
//	        type: string
//	        enum:
//	          - move
//	          - copy
//	      from:
//	        description: A JSON Pointer path.
//	        type: string
type PatchRequestBody struct{}
