.PHONY: clean lint test build \
		build-linux-arm64 build-linux-amd64 multi-arch-image-%  \
		start-local-db stop-local-db

BIN_NAME := reminderbot
MAIN_DIRECTORY := ./cmd

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
LOCAL_DB := $(shell docker ps -f "name=mongodb" --format '{{.Names}}')

# Default build target
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
DOCKER_BUILD_PLATFORMS ?= linux/amd64,linux/arm64

default: clean lint test build

start-local-db:
ifneq ($(LOCAL_DB),mongodb)
	docker run --restart unless-stopped -d --name mongo \
        -e MONGO_INITDB_ROOT_USERNAME=mongoadmin \
        -e MONGO_INITDB_ROOT_PASSWORD=secret \
        mongo:4.4
endif

stop-local-db:
	docker stop mongodb
	docker rm mongodb

lint:
	golangci-lint run

clean:
	rm -rf cover.out

test: clean
	go test -v -race -cover ./...

dist:
	mkdir dist

build: clean
	CGO_ENABLED=0 go build -v -o ${BIN_NAME} ${MAIN_DIRECTORY}

image: export GOOS := linux
image: export GOARCH := amd64
image: build
	docker build -t youkoulayley/$(BIN_NAME):latest .
