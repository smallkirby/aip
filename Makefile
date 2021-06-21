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

install: Makefile
	$(MAKE) build
	cp ./aip /usr/bin/aip

.PHONY: fmt build install run
