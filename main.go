package main

import (
	"embed"

	"todo-service/cmd"

	"github.com/tink3rlabs/magic/storage"
)

//go:generate go run build/generate.go
//go:embed config
var configFS embed.FS

func main() {
	storage.ConfigFs = configFS
	cmd.ConfigFS = configFS
	cmd.Execute()
}
