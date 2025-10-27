.PHONY: help install-tools dev build test clean run

help:
	@echo "make install-tools  - Install dev tools (Air)"
	@echo "make dev           - Start dev server with hot-reload"
	@echo "make build         - Build server binary"
	@echo "make test          - Run tests"
	@echo "make run           - Run server"
	@echo "make clean         - Clean artifacts"

install-tools:
	go install github.com/air-verse/air@latest

dev:
	@which air > /dev/null || (echo "Run: make install-tools" && exit 1)
	air

build:
	go build -o bin/wellknown-server ./cmd/server

test:
	go test -v ./pkg/...

clean:
	rm -rf tmp/ bin/

run: build
	./bin/wellknown-server
