
SHELL = /bin/sh

# define the source/target folders
SRC_ROOT = $(shell pwd)
CLIENT   = $(SRC_ROOT)/client
SERVER   = $(SRC_ROOT)/server

# Raspberry Pi binaries
PiScanner: $(CLIENT)/scanner.go
	go build -o $(CLIENT)/PiScanner $^

WebApp: $(CLIENT)/webapp.go
	go build -o $(CLIENT)/WebApp $^

PI_TARGETS = PiScanner WebApp

clients: $(PI_TARGETS)

# Server binary
APIServer: $(SERVER)/main.go
	go build -o $(SERVER)/APIServer $^

# all components
all: clients APIServer

clean:
	rm -f $(addprefix $(CLIENT)/, $(PI_TARGETS))
	rm -f $(addprefix $(SERVER)/, APIServer)
