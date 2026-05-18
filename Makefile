BINARY=ss
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build install test clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/ss/

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

test:
	go test ./... -v -race

clean:
	rm -rf bin/
