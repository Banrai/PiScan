// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/server/api"
	"net/http"
)

const (
	// API server definitions
	apiServer       = "saruzai.com"
	apiPort         = 9001
	apiSubdomain    = "api"
	apiSSL          = true
	apiExternalPort = 443

	// Coordinates for the POD db clone (mysql) on this server
	barcodeDBUser   = "pod"
	barcodeDBPass   = ""
	barcodeDBServer = "127.0.0.1"
	barcodeDBPort   = 3306
)

func main() {
	var (
		dbUser, dbPass, dbHost, host, subdomain string
		dbPort, port, externalPort              int
		useSSL                                  bool
	)

	flag.StringVar(&dbUser, "dbUser", barcodeDBUser, fmt.Sprintf("The barcodes database user (defaults to '%s')", barcodeDBUser))
	flag.StringVar(&dbPass, "dbPass", barcodeDBPass, fmt.Sprintf("The barcodes database password (defaults to '%s')", barcodeDBPass))
	flag.StringVar(&dbHost, "dbHost", barcodeDBServer, fmt.Sprintf("The barcodes database server (defaults to '%s')", barcodeDBServer))
	flag.IntVar(&dbPort, "dbPort", barcodeDBPort, fmt.Sprintf("The barcodes database port (defaults to '%d')", barcodeDBPort))
	flag.StringVar(&host, "host", apiServer, fmt.Sprintf("The hostname or IP address of the API server (defaults to '%s')", apiServer))
	flag.IntVar(&port, "port", apiPort, fmt.Sprintf("The internal API server port (defaults to '%d')", apiPort))
	flag.StringVar(&subdomain, "subdomain", apiSubdomain, fmt.Sprintf("The external subdomain of the API server (defaults to '%s')", apiSubdomain))
	flag.BoolVar(&useSSL, "ssl", apiSSL, fmt.Sprintf("Does the API server use SSL? (defaults to '%t')", apiSSL))
	flag.IntVar(&externalPort, "extPort", apiExternalPort, fmt.Sprintf("The external API server port (defaults to '%d')", apiExternalPort))
	flag.Parse()

	coords := api.DBConnection{Host: dbHost, User: dbUser, Pass: dbPass, Port: dbPort}

	// define the external-facing API server link
	// for email confirmations, etc.
	var buffer bytes.Buffer
	buffer.WriteString("http")
	if useSSL {
		buffer.WriteString("s")
	}
	buffer.WriteString("://")
	if len(subdomain) > 0 {
		buffer.WriteString(subdomain)
		buffer.WriteString(".")
	}
	buffer.WriteString(host)
	if !useSSL || externalPort != apiExternalPort {
		// the port matters only if it is non-standard
		// for ssl or if not using ssl at all
		buffer.WriteString(fmt.Sprintf(":%d", externalPort))
	}
	apiServerLink := buffer.String()

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
			return api.RegisterAccount(r, coords, apiServerLink)
		}
		api.Respond("application/json", "utf-8", register)(w, r)
	}

	// respond to the email verification link
	handlers["/verify/"] = func(w http.ResponseWriter, r *http.Request) {
		verify := func(w http.ResponseWriter, r *http.Request) string {
			return api.VerifyAccount(r, coords)
		}
		api.Respond("text/html", "utf-8", verify)(w, r)
	}

	// respond to account status requests
	handlers["/status"] = func(w http.ResponseWriter, r *http.Request) {
		register := func(w http.ResponseWriter, r *http.Request) string {
			return api.GetAccountStatus(r, coords)
		}
		api.Respond("application/json", "utf-8", register)(w, r)
	}

	// accept user-contributed data
	handlers["/contribute/"] = func(w http.ResponseWriter, r *http.Request) {
		contribute := func(w http.ResponseWriter, r *http.Request) string {
			return api.ContributeData(r, coords)
		}
		api.Respond("application/json", "utf-8", contribute)(w, r)
	}

	// email items list to a user
	handlers["/email/"] = func(w http.ResponseWriter, r *http.Request) {
		fn := func(w http.ResponseWriter, r *http.Request) string {
			return api.EmailSelectedItems(r, coords)
		}
		api.Respond("application/json", "utf-8", fn)(w, r)
	}

	api.NewAPIServer(host, api.DefaultServerTransport, port, api.DefaultServerReadTimeout, handlers)
}
