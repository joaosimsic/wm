.PHONY: build run clean fmt vet

BIN = ./wm

build:
	go build -o $(BIN) ./cmd/wm

run: build
	$(BIN)

dev: build
	go run ./cmd/wm-dev

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(BIN)
