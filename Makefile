.PHONY: build test clean

build:
	command go build -o bin/bt .

test:
	command go test ./...

clean:
	rm -f bin/bt
