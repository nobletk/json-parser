MAIN_PACKAGE_PATH := ./cmd/jsonparser
BINARY_NAME := jsonparser

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...

## test: run all tests
.PHONY: test
test:
	go test -v ./...

## build: build the application
.PHONY: build
build:
	# Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...
	go build -o=/tmp/bin/${binary_name} ${main_package_path}
