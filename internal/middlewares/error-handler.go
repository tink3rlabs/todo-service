package middlewares

import (
	"fmt"
	"log"
	"net/http"
	// "github.com/go-chi/chi/v5"
	// "github.com/go-chi/render"
)

type BaseError struct {
	Cause      error
	Code       int
	OwnMessage string
	StatusCode int
}

func NewBaseError(cause error, code int, ownMessage string, statusCode int) *BaseError {
	return &BaseError{
		Cause:      cause,
		Code:       code,
		OwnMessage: ownMessage,
		StatusCode: statusCode,
	}
}

func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.OwnMessage, e.Cause)
	}
	return e.OwnMessage
}

// TODO: Move to types
type ErrorResponse struct {
	Body       ErrorResponseBody
	StatusCode int
}

type ErrorResponseBody struct {
	Code      int
	Details   []ErrorDetails
	FieldName string
	Message   string
	Success   bool
}

type ErrorDetails struct {
	Code      int
	FieldName string
	Message   string
}

// func (e *BaseError) GetErrorDetails() []ErrorDetails {
// 	cause := e.Cause
// 	if cause == nil {
// 		return nil
// 	}
// }

func ApplicationErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Handling the error!")
		next.ServeHTTP(w, r)
	})
}
