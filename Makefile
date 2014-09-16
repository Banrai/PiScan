
GOFILES=$(wildcard *.go **/*.go)

format:
	gofmt -w=true ${GOFILES}

all:
	go build -o PiScanner main.go

clean:
	rm -rf PiScanner
