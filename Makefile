.PHONY: all build test docker-build run lint clean proto

all: lint test build

build:
	mkdir -p bin
	go build -ldflags="-w -s" -o bin/app ./cmd/app
	go build -ldflags="-w -s" -o bin/client ./cmd/client

test:
	go test -v -race ./...

docker-build:
	docker build -t app:latest .

run:
	go run ./cmd/app

lint:
	golangci-lint run ./...

clean:
	rm -rf bin coverage.out

proto:
	mkdir -p gen
	protoc --proto_path=proto \
		--go_out=gen --go_opt=module=github.com/kazah/golang-task-03/gen \
		--go-grpc_out=gen --go-grpc_opt=module=github.com/kazah/golang-task-03/gen \
		proto/rate/v1/rate.proto
