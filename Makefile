GO_MODULE_NAME = github.com/axatol/guosheng
BUILD_COMMIT ?= $(shell git rev-parse HEAD)
GO_BUILD_LDFLAGS = -X '$(GO_MODULE_NAME)/pkg/config.buildCommit=$(BUILD_COMMIT)'
GO_BUILD_LDFLAGS += -X '$(GO_MODULE_NAME)/pkg/config.buildTime=$(shell date +"%Y-%m-%dT%H:%M:%S%z")'

vet:
	go vet ./...

deps:
	go mod download

build:
	go build -o ./bin/guosheng -ldflags="$(GO_BUILD_LDFLAGS)" ./cmd/guosheng/main.go
