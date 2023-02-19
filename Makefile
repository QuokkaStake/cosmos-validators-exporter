VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
LDFLAGS = -X main.version=${VERSION}

build:
	go build -ldflags '$(LDFLAGS)' cmd/cosmos-validators-exporter.go

install:
	go install -ldflags '$(LDFLAGS)' cmd/cosmos-validators-exporter.go

lint:
	golangci-lint run --fix ./...
