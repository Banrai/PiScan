// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/server/api"
	"github.com/Banrai/PiScan/server/commerce"
	"github.com/Banrai/PiScan/server/commerce/amazon"
	"github.com/Banrai/PiScan/server/database/barcodes"
	"net/http"
	"strings"
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

// Lookup the barcode, using both the barcodes database, and the Amazon API
//func lookupBarcode(r *http.Request, statements map[string]*sql.Stmt) string {
func lookupBarcode(r *http.Request, db api.DBConnection) string {
	// the result is a json representation of the list of found products
	products := make([]*commerce.API, 0)

	// this function only responds to POST requests
	if "POST" == r.Method {
		r.ParseForm()

		barcodeVal, barcodeExists := r.PostForm["barcode"]
		if barcodeExists {
			statements := api.InitServerDatabase(db)
			asinLookup, asinLookupExists := statements[barcodes.ASIN_LOOKUP]
			asinInsert, asinInsertExists := statements[barcodes.ASIN_INSERT]
			if asinLookupExists && asinInsertExists {
				prods, prodErr := amazon.Lookup(strings.Join(barcodeVal, ""), asinLookup, asinInsert)
				if prodErr == nil {
					for _, prod := range prods {
						products = append(products, prod)
					}
				}
			}
		}
	}

	result, err := json.Marshal(products)
	if err != nil {
		fmt.Println(err)
	}
	return string(result)
}

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
	handlers["/lookup"] = func(w http.ResponseWriter, r *http.Request) {
		lookup := func(w http.ResponseWriter, r *http.Request) string {
			return lookupBarcode(r, coords)
		}
		api.Respond("application/json", "utf-8", lookup)(w, r)
	}
	api.NewAPIServer(host, port, api.DefaultServerReadTimeout, handlers)

}
