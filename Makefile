.PHONY: build test dev

export GO111MODULE=on

build:
	go build -o .build/remediator cmd/remediator/app.go

test: build
	go install github.com/grosser/go-testcov@latest
	go-testcov ./...
	go mod tidy && git diff --exit-code
	go fmt ./... && git diff --exit-code

dev: build
	.build/remediator
