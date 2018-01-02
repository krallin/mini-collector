GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
SHELL=/bin/bash

.PHONY: deps
deps:
	dep ensure

.PHONY: build
build: $(GOFILES_NOVENDOR)
	go list ./...  | grep cmd | xargs -P $$(nproc) -n 1 -- go build -i

writer/influxdb/api.proto.influxdb_formatter.go \
writer/datadog/api.proto.datadog_formatter.go \
publisher/api.proto.publisher_formatter.go: \
api/api.proto .codegen/emit.py \
.codegen/influxdb_formatter.go.jinja2 \
.codegen/datadog_formatter.go.jinja2 \
.codegen/publisher_formatter.go.jinja2
	protoc -I api api/api.proto --plugin=protoc-gen-custom=./.codegen/emit.py --custom_out=.
	find . -name "api.proto.*_formatter.go" | xargs gofmt -l -w

api/api.pb.go: api/api.proto
	protoc -I api/ api/api.proto --go_out=plugins=grpc:api

.PHONY: gofiles
src: $(GOFILES_NOVENDOR) fmt
	@true

.PHONY: unit
unit: $(GOFILES_NOVENDOR)
	go test $$(go list ./... | grep -v /vendor/)
	go vet $$(go list ./... | grep -v /vendor/)

.PHONY: test
test: unit
	@true

.PHONY: fmt
fmt:
	gofmt -l -w ${GOFILES_NOVENDOR}

.DEFAULT_GOAL := test
