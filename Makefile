BINARY := wget
PKG    := ./cmd

.PHONY: all build run test race cover vet fmt check clean help

all: check build

# Builds ./wget in the repo root, where the audit runs it.
build:
	go build -o $(BINARY) $(PKG)

# Usage: make run ARGS="-O=out.txt https://example.com/file"
run: build
	./$(BINARY) $(ARGS)

test:
	go test ./...

race:
	go test -race ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet:
	go vet ./...

fmt:
	gofmt -l -w .

check: fmt vet race

clean:
	rm -f $(BINARY) coverage.out

help:
	@echo "make build   build ./$(BINARY)"
	@echo "make run     build and run, pass flags with ARGS=\"...\""
	@echo "make test    run all tests"
	@echo "make race    run all tests with the race detector"
	@echo "make cover   run tests and show coverage per function"
	@echo "make vet     run go vet"
	@echo "make fmt     format all Go files"
	@echo "make check   fmt + vet + race"
	@echo "make clean   remove build and coverage output"
