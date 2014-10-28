
all:
	go build -o client/PiScanner client/scanner.go

clean:
	rm -rf client/PiScanner
