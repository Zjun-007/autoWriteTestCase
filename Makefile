.PHONY: build run test tidy

build:
	go build -o bin/awtc ./cmd/awtc

run: build
	./bin/awtc generate -i examples/login_requirement.md -o ./out -f json,excel,markdown

test:
	go test ./...

tidy:
	go mod tidy
