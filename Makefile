TARGET_PLATFORM ?= linux/amd64
TAG ?= latest
TEST_RUNNER ?= gotest.tools/gotestsum@latest

build:
	go build ./...

test:
	go run ${TEST_RUNNER} ./...

lint:
	golangci-lint run ./...

coverage:
	go run ${TEST_RUNNER} -- -coverprofile=cover.out ./...
	go tool cover -html cover.out -o cover.html

container-build:
	docker build . --platform=${TARGET_PLATFORM} -t szykol/http-go:${TAG}
