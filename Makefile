.PHONY: build
build:
	-go build -o ./tmp/app main.go
	
.PHONY: dev
dev:
	go build -o ./tmp/app main.go && air

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test:
	go test -race -v -timeout 30s ./...

.PHONY: release-notes
release-notes:
	echo "Releasing ...."
