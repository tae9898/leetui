GO ?= go
GOBIN ?= $(HOME)/go/bin

.PHONY: build test vet install clean

build:
	$(GO) build -o leetui ./cmd/leetui

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

install: build
	install -Dm755 leetui $(GOBIN)/leetui

clean:
	rm -f leetui
