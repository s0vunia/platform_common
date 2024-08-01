LOCAL_BIN:=$(CURDIR)/bin

lint:
	GOBIN=$(LOCAL_BIN) golangci-lint run ./... --config .golangci.pipeline.yaml

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

get-deps:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
