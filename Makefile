.PHONY: all test test-main test-cpp build run

all: test build

build:
	@go build ./bootstrap

test: test-main test-cpp

test-main:
	@go test ./bootstrap/... -count=1

test-cpp:
	@echo Running tests with C output
	@go test ./bootstrap -count=1 -args -bit.cpp

run: build
	@./bit --boot
