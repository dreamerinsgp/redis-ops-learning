.PHONY: build test

build:
	mkdir -p build && go build -o build/redis-ops ./cmd

test:
	go test ./...
