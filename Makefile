GO ?= go
GOBIN ?= $(HOME)/go/bin

.PHONY: build test vet install clean

build:
	$(GO) build -o pstui ./cmd/pstui

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

install: build
	install -Dm755 pstui $(GOBIN)/pstui

clean:
	rm -f pstui
