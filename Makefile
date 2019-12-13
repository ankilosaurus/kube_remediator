.PHONY: build test

export GO111MODULE=on

build:
	go build -o .build/remediator cmd/remediator/app.go

test: build
	go get github.com/grosser/go-testcov
	go-testcov ./...
	go mod tidy && git diff --exit-code
	go fmt ./... && git diff --exit-code

