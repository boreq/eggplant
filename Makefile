all: build

build:
	mkdir -p build
	go build -o ./build/goaccess ./main

build-race:
	mkdir -p build
	go build -race -o ./build/goaccess ./main

frontend:
	./_tools/build_frontend.sh

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/goaccess/"
	godoc -http=:6060

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf ./build

.PHONY: all build frontend build-race doc test test-verbose clean
