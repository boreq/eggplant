all: build

build:
	mkdir -p build
	go build -o ./build/plum ./cmd/plum

build-race:
	mkdir -p build
	go build -race -o ./build/plum ./cmd/plum

frontend:
	./_tools/build_frontend.sh

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/plum/"
	godoc -http=:6060

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf ./build

.PHONY: all build frontend build-race doc test test-verbose clean
