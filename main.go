package main

import (
	"embed"

	"todo-service/cmd"
	"todo-service/storage"
)

//go:embed config
var configFS embed.FS

func main() {
	storage.ConfigFs = configFS
	cmd.Execute()
}
