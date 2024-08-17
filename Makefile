TARGET_PLATFORM ?= linux/amd64
TAG ?= latest

build:
	go build ./...

test:
	go run gotest.tools/gotestsum@latest ./...

lint:
	golangci-lint run ./...

container-build:
	docker build . --platform=${TARGET_PLATFORM} -t szykol/http-go:${TAG}
