APP_NAME := webhook-service
CMD_DIR := ./cmd/$(APP_NAME)
OUTPUT_DIR := ./bin
OUTPUT := $(OUTPUT_DIR)/$(APP_NAME)

build:
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT) $(CMD_DIR)

run: build
	$(OUTPUT)

clean:
	rm -rf $(OUTPUT_DIR)

test:
	go test ./... -v

lint:
	go vet ./...

.PHONY: build run clean test lint