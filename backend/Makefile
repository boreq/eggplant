BUILD_DIRECTORY=_build
PROGRAM_NAME=eggplant

all: build

dev: test lint build

ci: tools dependencies generate tidy fmt check-repository-unchanged test lint build

build-directory:
	mkdir -p ./${BUILD_DIRECTORY}

build: build-directory
	go build -o ./${BUILD_DIRECTORY}/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

check-repository-unchanged:
	../tools/check_repository_unchanged.sh

tools:
	 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	 go install github.com/google/wire/cmd/wire@v0.7.0
	 go install golang.org/x/tools/cmd/goimports@v0.45.0

dependencies:
	go get ./...

tidy:
	go mod tidy

generate:
	 go generate ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf ./${BUILD_DIRECTORY}

fmt:
	goimports -w -l .

update-golden: build
	./_build/eggplant default_config > internal/config/data/config.golden.toml

.PHONY: all build build-directory check-repository-unchanged tools dependencies lint test test-verbose clean fmt tidy dev update-golden
