default: bin

# Build the project
.PHONY: bin
bin:
	go build ./cmd/...

# Run all available tests
.PHONY: test
test:
	go test ./...
