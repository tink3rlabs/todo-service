# TODO Service
TODO Service provides a framework for building a microservice in go-lang. 

## Developing

There are two ways you can develope

### Using a devcontainer

Using [devcontainers](https://containers.dev/) is the recommended approach as it ensures your development machine stays clean as well as provides consistency between different developers reducing the "works on my machine" problem.

To develope in a dev container follow these steps:

1. Install [Docker](https://www.docker.com/) and [VSCode](https://code.visualstudio.com/)
2. Install the [Dev Containers VSCode extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
3. Open the code base in vscode
4. You will be prompted to reopen the project in a dev container, follow the instructions

### Using the traditional approach

Install the [go language ](https://go.dev/) and your favorite IDE, then start developing using whichever method works best for you.

if you use Visual studio code, consider installing [Rich Go language support for Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=golang.Go) by the go team at Google. It provides several nifty features.

## Building and running
 ```bash
 go generate
```
 ```bash
 go build
```
 ```bash
 ./todo-service --config ./config/development.yaml server
```

## Testing with curl

### Creating a new TODO item

 ```bash
curl -X POST "http://localhost:8080/todos" \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{
      "summary": "New Todo item",
      "done": false
    }'
```

### Listing TODO items

```bash
curl http://localhost:8080/todos
```

### Getting a single TODO items

You can get the ID of the TODO item from either the response to the Create TODO API call, or the response to the List TODO API call

```bash
curl http://localhost:8080/todos/${TODO_ID}
```

### Updating a todo item (replace mode)

This method replaces all values of the TODO item with the specified ID with the ones provided in the request body.

You can get the ID of the TODO item from either the response to the Create TODO API call, or the response to the List TODO API call

```bash
curl -X PUT http://localhost:8080/${TODO_ID} \
  -H 'Content-Type: application/json' \
  -d '{"summary": "replaced", "done": true}'
```

### Updating a todo item (patch mode)

This method folows the [JSONPatch](https://jsonpatch.com/) format to update specific values of the todo with the specified ID with the ones provided in the request body.

You can get the ID of the TODO item from either the response to the Create TODO API call, or the response to the List TODO API call

```bash
curl -X PATCH http://localhost:8080/todos/${TODO_ID} \
     -H 'accept: application/json' \
     -H 'Content-Type: application/json-patch+json' \
     -d '[
       {
         "op": "replace",
         "path": "/summary",
         "value": "patched"
       },
       {
         "op": "replace",
         "path": "/done",
         "value": true
       }
     ]'
```

### Deleting a single TODO items

You can get the ID of the TODO item from either the response to the Create TODO API call, or the response to the List TODO API call

```bash
curl -X DELETE http://localhost:8080/todos/${TODO_ID}
```
