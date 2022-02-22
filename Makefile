ARCH ?= $(shell go env GOARCH)

.PHONY: fmt
fmt: ## Run go fmt -s all over the project
	@gofmt -s -w $(shell find . -name "*.go")

.PHONY: protobuf-build
protobuf-build: ## Build protocol buffers into model definitions
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/api.proto

.PHONY: protobuf-fmt
protobuf-fmt: ## Clean and format protocol buffer files
	prototool format -w api/api.proto

.PHONY: build
build: ## Build all applications and cmd utitilies
	GO_BUILD_FLAGS="-v" ./build.sh

# source: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
