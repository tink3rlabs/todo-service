package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"todo-service/internal/storage"
	"todo-service/types"
)

type Responder struct{}

func (f *Responder) Respond(object interface{}, err error, w http.ResponseWriter, r *http.Request, code int) {

	// No content
	if object == nil && err == nil {
		render.Status(r, code)
		render.NoContent(w, r)
		return
	}

	if errors.Is(err, storage.ErrNotFound) {
		render.Status(r, 404)
		response := types.ErrorResponse{
			Status: "NOT_FOUND",
			Error:  "Todo not found",
		}
		render.JSON(w, r, response)
		return

	}
	if errors.Is(err, storage.ValidationError) {
		render.Status(r, 400)
		response := types
	}
	// Any other error
	if err != nil {
		render.Status(r, 500)
		response := types.ErrorResponse{
			Status: "SERVER_ERROR",
			Error:  "Encountered an unexpected server error: " + err.Error(),
		}
		render.JSON(w, r, response)
		return
	}

	render.Status(r, code)
	render.JSON(w, r, object)
}

// TODO: Refine errors and add swagger declarative comments
// Response is a wrapper response structure
// Response example
type Response struct {
	Status interface{} `json:"status"`
	Data   interface{} `json:"data"`
}

// ResponseMeta example
type ResponseMeta struct {
	AppStatusCode int    `json:"code"`
	Message       string `json:"statusType,omitempty"`
	ErrorDetail   string `json:"errorDetail,omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
	DevMessage    string `json:"devErrorMessage,omitempty"`
}

// ErrResponse example
type ErrResponse struct {
	HTTPStatusCode int          `json:"-"` // http response status code
	Status         ResponseMeta `json:"status"`
	AppCode        int64        `json:"code,omitempty"` // application-specific error code
}

// HTTPErr example
type HTTPErr struct {
	Err  error
	Code int
}

func (e HTTPErr) Error() string {
	return fmt.Sprintf("%s", e.Err)
}

// New returns new http error from error object and a code
func New(err error, code int) *HTTPErr {
	return &HTTPErr{
		Err:  err,
		Code: code,
	}
}

// Error example
// Error returns an HTTPError
func Error(err interface{}) *HTTPErr {
	switch err.(type) {
	case *HTTPErr:
		return err.(*HTTPErr)
	case error:
		return &HTTPErr{
			Err:  err.(error),
			Code: http.StatusInternalServerError,
		}
	default:
		return &HTTPErr{
			Err:  errors.New("Unknown error"),
			Code: http.StatusInternalServerError,
		}
	}
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// Render for All Responses
func (rd *Response) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func WrapHandlerFunc(route string, name string, handler http.HandlerFunc) (string, http.HandlerFunc) {
	return route, handler
}

// NewSuccessResponse returns a new success response
func NewSuccessResponse(status int, data interface{}) *Response {
	return &Response{
		Status: &ResponseMeta{
			AppStatusCode: status,
			Message:       "SUCCESS",
		},
		Data: data,
	}
}

// ErrInvalidRequest returns a 400
func ErrInvalidRequest(err error, message string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		Status: ResponseMeta{
			AppStatusCode: http.StatusBadRequest,
			Message:       "ERROR",
			ErrorMessage:  "Invalid Request",
			ErrorDetail:   message,
			DevMessage:    err.Error(),
		},
	}
}

// ErrNotFound returns a 404
var ErrNotFound = &ErrResponse{
	HTTPStatusCode: http.StatusNotFound,
	Status: ResponseMeta{
		AppStatusCode: http.StatusNotFound,
		Message:       "ERROR",
		ErrorDetail:   "Resource not found",
		ErrorMessage:  "The endpoint you were seeking burned down",
	},
}
