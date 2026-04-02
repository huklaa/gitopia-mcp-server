.PHONY: build test lint docker clean

VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o server ./cmd/server

test:
	go test -race -coverprofile=coverage.out ./...

lint:
	go vet ./...
	golangci-lint run

docker:
	docker build -t gitopia-mcp-server .

clean:
	rm -f server coverage.out
