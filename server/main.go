// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/server/api"
	"net/http"
)

const (
	// API server definitions
	apiServer = ""
	apiPort   = 9001

	// Coordinates for the POD db clone (mysql) on this server
	barcodeDBUser   = "pod"
	barcodeDBPass   = ""
	barcodeDBServer = "127.0.0.1"
	barcodeDBPort   = 3306
)

func main() {
	var (
		dbUser, dbPass, dbHost, host string
		dbPort, port                 int
	)

	flag.StringVar(&dbUser, "dbUser", barcodeDBUser, fmt.Sprintf("The barcodes database user (defaults to '%s')", barcodeDBUser))
	flag.StringVar(&dbPass, "dbPass", barcodeDBPass, fmt.Sprintf("The barcodes database password (defaults to '%s')", barcodeDBPass))
	flag.StringVar(&dbHost, "dbHost", barcodeDBServer, fmt.Sprintf("The barcodes database server (defaults to '%s')", barcodeDBServer))
	flag.IntVar(&dbPort, "dbPort", barcodeDBPort, fmt.Sprintf("The barcodes database port (defaults to '%d')", barcodeDBPort))
	flag.StringVar(&host, "host", apiServer, fmt.Sprintf("The hostname or IP address of the API server (defaults to '%s')", apiServer))
	flag.IntVar(&port, "port", apiPort, fmt.Sprintf("The API server port (defaults to '%d')", apiPort))
	flag.Parse()

	coords := api.DBConnection{Host: dbHost, User: dbUser, Pass: dbPass, Port: dbPort}

	handlers := map[string]func(http.ResponseWriter, *http.Request){}

	// respond to a barcode lookup request
	handlers["/lookup"] = func(w http.ResponseWriter, r *http.Request) {
		lookup := func(w http.ResponseWriter, r *http.Request) string {
			return api.LookupBarcode(r, coords)
		}
		api.Respond("application/json", "utf-8", lookup)(w, r)
	}

	// respond to contributor account creation requests
	handlers["/register"] = func(w http.ResponseWriter, r *http.Request) {
		register := func(w http.ResponseWriter, r *http.Request) string {
			return api.RegisterAccount(r, coords, host, port)
		}
		api.Respond("application/json", "utf-8", register)(w, r)
	}

	api.NewAPIServer(host, port, api.DefaultServerReadTimeout, handlers)
}
