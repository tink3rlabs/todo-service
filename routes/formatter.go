package routes

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"todo-service/internal/storage"
	"todo-service/types"
)

type Formatter struct{}

func (f *Formatter) Respond(object interface{}, err error, w http.ResponseWriter, r *http.Request) {
	if errors.Is(err, storage.ErrNotFound) {
		render.Status(r, 404)
		response := types.ErrorResponse{
			Status: "NOT_FOUND",
			Error:  "Todo not found",
		}
		render.JSON(w, r, response)
		return
	}

	if err != nil {
		render.Status(r, 500)
		response := types.ErrorResponse{
			Status: "SERVER_ERROR",
			Error:  "Encountered an unexpected server error: " + err.Error(),
		}
		render.JSON(w, r, response)
		return
	}

	render.JSON(w, r, object)
}
