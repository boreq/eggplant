all: build

build:
	mkdir -p build
	go build -o ./build/goaccess ./main

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/goaccess/"
	godoc -http=:6060

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf ./build

.PHONY: all build doc test test-verbose clean
