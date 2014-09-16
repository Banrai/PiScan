GOFMT=gofmt -s -tabs=false -tabwidth=4

GOFILES=$(wildcard *.go **/*.go)

format:
	${GOFMT} -w ${GOFILES}

all:
	go build -o PiScanner main.go

clean:
	rm -rf PiScanner
