# TODO Service
TODO Service provides a framework for building a microservice in go-lang. 

## Source code editing
Please install visual studio code and the go language. The version used can be seen in the go.mod file. Also consider installing "Rich Go language support for Visual Studio Code" by the go team at Google. It provides several nifty features
https://marketplace.visualstudio.com/items?itemName=golang.Go


## Running in a container
To run the service in a container, please install Visual Studio Code. Setup and configuration of visual studio devcontainers is discussed in detail below
https://code.visualstudio.com/docs/devcontainers/containers
The file devcontainer.json file has more relevant pointers and the json file can be viewed through Visual Studio Code.

When the container is running one can see the Dev Container: Go @ xxxx in the bottom left hand corner of Visual Studio Code. One can see more details about the container in Docker Desktop or Orbstack https://orbstack.dev/ which is a newer alternative to Docker Desktop.

## Code organization

This section goes through the code organization of todo microservice.

Each target in the Makefile helps automate a specific development task,streamlining the workflow for building, testing, and analyzing the Go application.

main.go is the entry point to the micro-service.

### cmd package
The cmd package files are in the cmd folder. This uses cobra and viper packages to handle initialization and configuration info. The runServer cmd sets up the API routes for the service and checks the api-docs, health/liveness and health/readiness endpoints. 


### todo package in features folder
The features specific to a microservice are to be implemented in the todo package. The code is in the features folder in todo package

### routes package
Routes package has sample implementations for get, delete and post routes. 

### types package
This package has ErrorResponse and ToDo schema used elsewhere in the microservice

### health package
This is in the internal/health folder. Implements healthchecker for the microservice

### leadership package
This is in the internal/leadership folder. Implements leader election from a list of cluster members and has the related helper functions.

### middlewares package
This is in the internal/middlewares folder. This validates the json schema and calls the next handler in the http chain to process the request further.

### storage package
This is in the internal/storage folder. This provides different storage adapters such as "memory", "sql"



