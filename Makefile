TARGET_PLATFORM ?= linux/amd64
TAG ?= latest

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

container-build:
	docker build . --platform=${TARGET_PLATFORM} -t szykol/http-go:${TAG}
