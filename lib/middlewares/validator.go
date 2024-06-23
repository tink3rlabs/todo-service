package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/xeipuuv/gojsonschema"
)

type Validator struct{}

type ValidationError struct {
	Status string   `json:"status"`
	Error  []string `json:"error"`
}

// ValidateJSON validates the JSON data against the provided  JSON schema
// https://pkg.go.dev/github.com/xeipuuv/gojsonschema#section-readme

func JSONSchemaValidator(schemaLoader gojsonschema.JSONLoader, data interface{}) (bool, []string) {
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return false, []string{err.Error()}
	}

	if result.Valid() {
		return true, nil
	}

	var errors []string
	for _, desc := range result.Errors() {
		errors = append(errors, desc.String())
	}

	return false, errors
}

func (f *Validator) ValidateRequest(schemas map[string]string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errors []string

		for target, schema := range schemas {
			schemaLoader := gojsonschema.NewStringLoader(schema)
			var data interface{}

			switch target {
			case "body":
				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					errors = append(errors, fmt.Sprintf("Invalid JSON body: %v", err))
					continue
				}
			case "query":
				queryData := make(map[string]interface{})
				for key, values := range r.URL.Query() {
					if len(values) > 0 {
						queryData[key] = values[0]
					}
				}
				data = queryData
			case "params":
				params := make(map[string]interface{})
				routeCtx := chi.RouteContext(r.Context())
				for i, key := range routeCtx.URLParams.Keys {
					params[key] = routeCtx.URLParams.Values[i]
				}
				data = params
			}

			isValid, validationErrors := JSONSchemaValidator(schemaLoader, data)
			if !isValid {
				errors = append(errors, validationErrors...)
			}
		}

		if len(errors) > 0 {
			validationError := ValidationError{
				Status: "INVALID_RESOURCE",
				Error:  errors,
			}
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, validationError)
			return
		}

		next.ServeHTTP(w, r)
	}
}
