.PHONY: build clean test

GO=go
GOOS?=linux
GOARCH?=amd64

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o bin/frachi src/main.go

clean:
	rm -rf bin/

test:
	$(GO) test ./...

all: clean build 