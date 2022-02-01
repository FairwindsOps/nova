# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=nova
COMMIT := $(shell git rev-parse HEAD)
VERSION := "local-dev"

all: lint test
build:
	$(GOBUILD) -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -s -w" -v
lint:
	golangci-lint run
reportcard:
	goreportcard-cli -t 100 -v
test:
	$(GOCMD) test -v --bench --benchmem -coverprofile coverage.txt -covermode=atomic ./...
	$(GOCMD) vet ./... 2> govet-report.out
	$(GOCMD) tool cover -html=coverage.txt -o cover-report.html
	printf "\nCoverage report available at cover-report.html\n\n"
tidy:
	$(GOCMD) mod tidy -compat=1.17
clean:
	$(GOCLEAN)
	$(GOCMD) fmt ./...
	rm -f $(BINARY_NAME)
# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -s -w" -v
build-docker: build-linux
	docker build -t quay.io/fairwinds/$(BINARY_NAME):dev .
