GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
SHELL=/bin/bash

.PHONY: deps
deps:
	dep ensure

.PHONY: build
build: $(GOFILES)
	go build

.PHONY: unit
unit:
	go test $$(go list ./... | grep -v /vendor/)
	go vet $$(go list ./... | grep -v /vendor/)

.PHONY: test
test: unit
	true

.PHONY: fmt
fmt:
	gofmt -l -w ${GOFILES_NOVENDOR}

.DEFAULT_GOAL := test
