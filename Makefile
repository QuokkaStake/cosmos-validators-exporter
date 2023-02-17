build:
	go build cmd/cosmos-validators-exporter.go

install:
	go install cmd/cosmos-validators-exporter.go

lint:
	golangci-lint run --fix ./...