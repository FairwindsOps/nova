# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
BINARY_NAME=nova
COMMIT := $(shell git rev-parse HEAD)
VERSION := "local-dev"

all: lint test
build:
	pkger
	$(GOBUILD) -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -s -w" -v