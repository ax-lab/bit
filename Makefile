.PHONY: all test build

all: test build

build:
	@go build ./bootstrap

test:
	@go test ./bootstrap/...
