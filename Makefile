.PHONY: build build-linux build-windows build-all

CGO_CFLAGS = -Wno-discarded-qualifiers

build: build-all

build-linux:
	CGO_CFLAGS=$(CGO_CFLAGS) go build -o bin/survival-bot-linux ./cmd/bot

build-windows:
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) GOOS=windows GOARCH=amd64 go build -o bin/survival-bot.exe ./cmd/bot

build-all: build-linux build-windows

test: tidy
	mockery
	CGO_CFLAGS=$(CGO_CFLAGS) go test ./...

tidy:
	go mod tidy
	go fmt ./...
