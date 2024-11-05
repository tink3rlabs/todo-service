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

## Building and running
 ```bash
 go generate
```
 ```bash
 go install
```
 ```bash
 go build
```
 ```bash
 ./todo-service --config ./config/development.yaml server
```

## Testing with curl
 ```bash
 curl -X POST "http://localhost:8080/todos" \           
     -H "accept: application/json" \
     -H "Content-Type: application/json" \
     -d '{
       "summary": "New Todo item",
       "done": false
     }'
```
```bash
 curl -v http://localhost:8080/todos
```
```bash
curl -X PATCH "http://localhost:8080/todos/0192fa03-b02a-78af-a4e2-255326f5d891" \          
     -H "accept: application/json" \                 
     -H "Content-Type: application/json-patch+json" \
     -d '[
       {                 
         "op": "replace",   
         "path": "/summary",                    
         "value": "An updated TODO item summary"
       },
       {                 
         "op": "replace",
         "path": "/done",
         "value": true
       }
     ]'
```
```bash
curl -X DELETE "http://localhost:8080/todos/0192fa08-a84c-7d83-87aa-7c31ec29aacf" -H "accept: application/json"
```


