
SHELL = /bin/sh

SRC_ROOT = $(shell pwd)
CLIENT   = $(SRC_ROOT)/client

PiScanner: $(CLIENT)/scanner.go
	go build -o $(CLIENT)/PiScanner $^

WebApp: $(CLIENT)/webapp.go
	go build -o $(CLIENT)/WebApp $^

PI_TARGETS = PiScanner WebApp

clients: $(PI_TARGETS)

all: clients

clean:
	rm -f $(addprefix $(CLIENT)/, $(PI_TARGETS))
