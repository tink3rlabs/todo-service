package middlewares

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	serviceErrors "todo-service/internal/errors"
	"todo-service/internal/storage"
	"todo-service/internal/types"
)

type ErrorHandler struct{}

func (e *ErrorHandler) Wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var notFoundError *serviceErrors.NotFound
		var badRequestError *serviceErrors.BadRequest
		var serviceUnavailable *serviceErrors.ServiceUnavailable

		err := handler(w, r)

		if (errors.As(err, &notFoundError)) || (errors.Is(err, storage.ErrNotFound)) {
			render.Status(r, http.StatusNotFound)
			response := types.ErrorResponse{
				Status: http.StatusText(http.StatusNotFound),
				Error:  err.Error(),
			}
			render.JSON(w, r, response)
			return
		}

		if errors.As(err, &badRequestError) {
			render.Status(r, http.StatusBadRequest)
			response := types.ErrorResponse{
				Status: http.StatusText(http.StatusBadRequest),
				Error:  err.Error(),
			}
			render.JSON(w, r, response)
			return
		}

		if errors.As(err, &serviceUnavailable) {
			render.Status(r, http.StatusServiceUnavailable)
			response := types.ErrorResponse{
				Status: http.StatusText(http.StatusServiceUnavailable),
				Error:  err.Error(),
			}
			render.JSON(w, r, response)
			return
		}

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			response := types.ErrorResponse{
				Status: http.StatusText(http.StatusInternalServerError),
				Error:  "Encountered an unexpected server error: " + err.Error(),
			}
			render.JSON(w, r, response)
			return
		}
	}
}
