BINARY     := cctv
MODULE     := github.com/lyarwood/cctv
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X '$(MODULE)/internal/cmd.Version=$(VERSION)'"
GO         := go
GINKGO     := ginkgo

.PHONY: all build test test-verbose test-cover lint fmt vet clean install run demo

all: build

build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY) ./cmd/cctv

test:
	$(GINKGO) run -r ./internal/...

test-verbose:
	$(GINKGO) run -r -v ./internal/...

test-cover:
	$(GINKGO) run -r --coverprofile=coverage.out ./internal/...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf bin/ coverage.out coverage.html

install: build
	install -m 0755 bin/$(BINARY) $(shell $(GO) env GOPATH)/bin/$(BINARY)

run: build
	./bin/$(BINARY)

demo: build
	@mkdir -p /tmp/cctv-demo/projects/demo-project /tmp/cctv-demo/sessions
	@cp demo/sessions-index.json /tmp/cctv-demo/projects/demo-project/
	@cp demo/*.jsonl /tmp/cctv-demo/projects/demo-project/
	vhs demo.tape
	@rm -rf /tmp/cctv-demo
