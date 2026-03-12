.PHONY: build test lint tidy migrate-up

build:
	go build ./...

test:
	go test ./...

tidy:
	go work sync

lint:
	golangci-lint run ./...

migrate-up:
	@echo "TODO: ..."