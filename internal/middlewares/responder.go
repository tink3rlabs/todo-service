package middlewares

import (
	"errors"
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

	// // Switch errors
	// switch t := err.(type) {
	// default:
	// 	fmt.Printf("Don't know type %T\n", t)
	// }

	if errors.Is(err, storage.ErrNotFound) {
		render.Status(r, 404)
		response := types.ErrorResponse{
			Status: "NOT_FOUND",
			Error:  "Todo not found",
		}
		render.JSON(w, r, response)
		return

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
