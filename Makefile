.PHONY: build test

build:
	@go run maker.go -v
	@ln -s -f maker bit

test:
	@go test ./boot/...

clean:
	@rm -rf ./build
	@rm -f maker
	@rm -f bit
