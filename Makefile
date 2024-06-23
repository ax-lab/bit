.PHONY: build test clean

build:
	@go build -o .build/ ./boot/cmd/bit.go

test:
	@go test ./boot/... -count=1

clean:
	@rm -rf .build
