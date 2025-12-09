.PHONY: build run clean test

APP_NAME=gomentum

build:
	go build -o $(APP_NAME).exe ./cmd/gomentum

run: build
	./$(APP_NAME).exe

clean:
	del $(APP_NAME).exe
	del $(APP_NAME).db

test:
	go test ./...
