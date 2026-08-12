BUILD_DIRECTORY=_build
PROGRAM_NAME=eggplant

.PHONY: all
all: build-without-frontend

.PHONY: dev
dev: test lint build-without-frontend

.PHONY: ci
ci: tools dependencies generate tidy fmt check-repository-unchanged test lint build-without-frontend

.PHONY: build-directory
build-directory:
	mkdir -p ./${BUILD_DIRECTORY}

.PHONY: build-without-frontend
build-without-frontend: build-directory
	CGO_ENABLED=0 go build -o ./${BUILD_DIRECTORY}/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

.PHONY: build-with-frontend
build-with-frontend: build-directory
	CGO_ENABLED=0 go build -tags withfrontend \
		-ldflags "-X github.com/boreq/eggplant/internal/version.Backend=${VERSION} -X github.com/boreq/eggplant/internal/version.Frontend=${FRONTEND_VERSION}" \
		-o ./${BUILD_DIRECTORY}/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

.PHONY: run
run:
	go run ./cmd/${PROGRAM_NAME} run --verbosity debug ${CONFIG}

.PHONY: check-repository-unchanged
check-repository-unchanged:
	../tools/check_repository_unchanged.sh

.PHONY: tools
tools:
	 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	 go install github.com/google/wire/cmd/wire@v0.7.0
	 go install golang.org/x/tools/cmd/goimports@v0.45.0

.PHONY: dependencies
dependencies:
	go get ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: generate
generate:
	 go generate ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf ./${BUILD_DIRECTORY}

.PHONY: fmt
fmt:
	goimports -w -l .

.PHONY: update-golden
update-golden: build-without-frontend
	./_build/eggplant default_config > internal/config/data/config.golden.toml
