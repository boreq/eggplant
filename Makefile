BUILD_DIRECTORY=_build
PROGRAM_NAME=eggplant

all: test lint build

build-directory:
	mkdir -p ./${BUILD_DIRECTORY}

build: build-directory
	go build -o ./${BUILD_DIRECTORY}/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

build-race: build-directory
	go build -race -o ./${BUILD_DIRECTORY}/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

frontend:
	./_tools/build_frontend.sh

tools:
	 go get -u honnef.co/go/tools/cmd/staticcheck
	 go get -u github.com/google/wire/cmd/wire

generate:
	 go generate ./...

lint: 
	go vet ./...
	staticcheck ./...

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/${PROGRAM_NAME}/"
	godoc -http=:6060

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf ./${BUILD_DIRECTORY}

.PHONY: all build build-directory frontend build-race tools lint doc test test-verbose clean
