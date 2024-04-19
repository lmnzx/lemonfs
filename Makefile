build:
	@go build -o bin/lemonfs

run: build
	@./bin/lemonfs

test:
	@go test ./...	-v
