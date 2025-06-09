.PHONY: build clean

GO=go
GOOS?=linux
GOARCH?=amd64

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o bin/frachi src/main.go

clean:
	rm -rf bin/

all: clean build 