// --8<-- [start:main-run]
package main

import (
	"embed"

	"todo-service/cmd"

	"github.com/tink3rlabs/magic/storage"
)

//go:generate go run build/generate.go
// --8<-- [start:main-imports]
//go:embed config
var configFS embed.FS

func main() {
	storage.ConfigFs = configFS
	// --8<-- [end:main-imports]
	cmd.ConfigFS = configFS
	cmd.Execute()
}
// --8<-- [end:main-run]
