package main

import (
	"encoding/json"
	"log"
	"net/http"
	"todo-service/routes"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	openapigodoc "github.com/tink3rlabs/openapi-godoc"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.Logger,          // Log API request calls
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	)

	router.Route("/", func(r chi.Router) {
		r.Mount("/todos", routes.TodoRoutes())
	})

	return router
}

func GenerateOpenApiSpec() []byte {
	securitySchemasData := []byte(`
	{
		"petstore_auth": {
			"type": "oauth2",
			"flows": {
				"implicit": {
					"authorizationUrl": "https://petstore3.swagger.io/oauth/authorize",
					"scopes": {
						"write:pets": "modify pets in your account",
						"read:pets": "read your pets"
					}
				}
			}
		},
		"api_key": {
			"type": "apiKey",
			"name": "api_key",
			"in": "header"
		}
	}`)

	var securitySchemas map[string]interface{}
	err := json.Unmarshal(securitySchemasData, &securitySchemas)
	if err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	apiDefinition := openapigodoc.OpenAPIDefinition{
		OpenAPI: "3.0.3",
		Info: openapigodoc.Info{
			Title:       "Todo API",
			Version:     "1.0.0",
			Description: "Simple example Todo API",
			Contact:     &openapi3.Contact{Email: "developer@example.com"},
			License:     &openapi3.License{Name: "Apache 2.0", URL: "http://www.apache.org/licenses/LICENSE-2.0.html"},
		},
		Servers: []openapigodoc.Server{{URL: "http://localhost:8080"}},
		Tags: []openapigodoc.Tag{
			{
				Name:         "todos",
				Description:  "Manage Todo items",
				ExternalDocs: &openapi3.ExternalDocs{URL: "http://example.com", Description: "Find out more"},
			},
		},
		ExternalDocs: openapigodoc.ExternalDocs{Description: "Find out more", URL: "http://example.com"},
		Components: openapigodoc.Components{
			SecuritySchemes: securitySchemas,
		},
	}

	definition, err := openapigodoc.GenerateOpenApiDoc(apiDefinition)
	if err != nil {
		log.Panicf("Logging err: %s\n", err.Error())
	}
	return definition
}

func main() {
	openApiSpec := GenerateOpenApiSpec()
	router := Routes()

	router.Get("/api-docs", func(w http.ResponseWriter, r *http.Request) {
		w.Write(openApiSpec)
	})

	log.Fatal(http.ListenAndServe(":8080", router)) // Note, the port is usually gotten from the environment.
}
