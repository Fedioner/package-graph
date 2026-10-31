.PHONY: build run test

build:
	go build -o bin/package-graph ./src

run:
	go run ./src

test:
	go test -v ./...
