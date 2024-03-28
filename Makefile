.PHONY: build test test-integration generate install-paramgen install-tools golangci-lint-install

VERSION=$(shell git describe --tags --dirty --always)

build:
	go build -ldflags "-X 'github.com/alarbada/conduit-connector-activemq-artemis.version=${VERSION}'" -o conduit-connector-activemq-artemis cmd/connector/main.go

test: up
	go test -v -v -race ./...; ret=$$?; \
		docker compose -f test/docker-compose.yml down; \
		exit $$ret

generate:
	go generate ./...

install-paramgen:
	go install github.com/conduitio/conduit-connector-sdk/cmd/paramgen@latest

install-tools:
	@echo Installing tools from tools.go
	@go list -e -f '{{ join .Imports "\n" }}' tools.go | xargs -tI % go install %
	@go mod tidy

lint:
	golangci-lint run

up:
	docker compose -f test/docker-compose.yml up --quiet-pull -d --wait 

down:
	docker compose -f test/docker-compose.yml down --volumes --remove-orphans
