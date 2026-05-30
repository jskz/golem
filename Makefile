GO ?= go
BIN ?= golem
MAIN_PKG := ./src
CGO_ENABLED ?= 0

.DEFAULT_GOAL := build

.PHONY: all build run test fmt clean

all: test build

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BIN) $(MAIN_PKG)

run: build
	./$(BIN)

test:
	$(GO) test ./...

fmt:
	gofmt -w src/*.go

clean:
	rm -f $(BIN)
