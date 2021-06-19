GOCMD=go
GOTEST=$(GOCMD) test -v
GOBUILD=$(GOCMD) build
BINARY_NAME=aip

run:
	$(GOCMD) run main.go

build:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)

fmt: 
	find . -type f -name "*.go" | xargs -i $(GOCMD) fmt {}

.PHONY: fmt build