BUILD_COMMIT ?= $(shell git rev-parse HEAD)
BUILD_TIME ?= $(shell date +"%Y-%m-%dT%H:%M:%S%z")
GO_MODULE_NAME = github.com/axatol/guosheng
GO_BUILD_LDFLAGS = -X '$(GO_MODULE_NAME)/pkg/config.buildCommit=$(BUILD_COMMIT)'
GO_BUILD_LDFLAGS += -X '$(GO_MODULE_NAME)/pkg/config.buildTime=$(BUILD_TIME)'

vet:
	go vet ./...

deps:
	go mod download

build:
	go build -o ./bin/ -ldflags="$(GO_BUILD_LDFLAGS)" ./...
