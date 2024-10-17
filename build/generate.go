//go:build tools

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tink3rlabs/magic/types"
	openapigodoc "github.com/tink3rlabs/openapi-godoc"
)

func generateOApiSpec() ([]byte, error) {
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
		return nil, err
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

	definition, err := openapigodoc.GenerateOpenApiDoc(apiDefinition, false)
	if err != nil {
		return nil, err
	}
	return definition, nil
}

func main() {
	fmt.Println("Generating OpenAPI definition file")
	path := "./config/openapi.json"

	fmt.Println("Generating local OpenAPI definition")
	openApiSpec, err := generateOApiSpec()
	if err != nil {
		fmt.Printf("error generating OpenAPI definition: %v\n", err)
	}

	fmt.Println("Merginig local OpenAPI definition with tink3rlabs magic definition")
	finalDefinition, err := types.MergeOpenAPIDefinitions(openApiSpec)
	if err != nil {
		fmt.Printf("error merging OpenAPI definition with common definitions: %v\n", err)
	}

	fmt.Println("Validating OpenAPI definition")
	valid, err := openapigodoc.ValidateOpenApiDoc(finalDefinition)
	if valid {
		fmt.Println("OpenAPI definition validated successfully")
	} else {
		fmt.Printf("error validating OpenAPI definition: %v\n", err)
		return
	}

	err = os.WriteFile(path, finalDefinition, 0644)
	if err != nil {
		fmt.Printf("error writing OpenAPI definition to file: %v", err)
	}

	fmt.Printf("OpenAPI definition was written to %s", path)
}
