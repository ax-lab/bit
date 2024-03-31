.PHONY: test bench

test:
	@go test ./...

bench:
	@go test -benchmem ./... -v -run ^\$$  -bench "^Bench"
