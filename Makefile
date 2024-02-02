.PHONY: all run test build

all: test build

run:
	@go run ./bootstrap

build:
	@go build ./bootstrap

test:
	@go test ./bootstrap/...
