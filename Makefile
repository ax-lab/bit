.PHONY: all test

all:
	@./bit sample/hello.bit

test:
	@go test ./boot/... -count=1
