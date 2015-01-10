// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// This is a fully-functional (but simple) PiScanner application.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/client/database"
	"github.com/Banrai/PiScan/scanner"
	"github.com/Banrai/PiScan/server/commerce"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	apiServerHost = "https://api.saruzai.com"
	apiServerPort = 9001
)

func main() {
	var (
		device, apiServer, sqlitePath, sqliteFile, sqliteTablesDefinitionPath string
		apiPort                                                               int
	)

	flag.StringVar(&device, "device", scanner.SCANNER_DEVICE, fmt.Sprintf("The '/dev/input/event' device associated with your scanner (defaults to '%s')", scanner.SCANNER_DEVICE))
	flag.StringVar(&apiServer, "apiHost", apiServerHost, fmt.Sprintf("The hostname or IP address of the API server (defaults to '%s')", apiServerHost))
	flag.IntVar(&apiPort, "apiPort", apiServerPort, fmt.Sprintf("The API server port (defaults to '%d')", apiServerPort))
	flag.StringVar(&sqlitePath, "sqlitePath", database.SQLITE_PATH, fmt.Sprintf("Path to the sqlite file (defaults to '%s')", database.SQLITE_PATH))
	flag.StringVar(&sqliteFile, "sqliteFile", database.SQLITE_FILE, fmt.Sprintf("The sqlite database file (defaults to '%s')", database.SQLITE_FILE))
	flag.StringVar(&sqliteTablesDefinitionPath, "sqliteTables", "", fmt.Sprintf("Path to the sqlite database definitions file, %s, (use only if creating the client db for the first time)", database.TABLE_SQL_DEFINITIONS))
	flag.Parse()

	if len(sqliteTablesDefinitionPath) > 0 {
		// this is a request to create the client db for the first time
		initDb, initErr := database.InitializeDB(database.ConnCoordinates{sqlitePath, sqliteFile, sqliteTablesDefinitionPath})
		if initErr != nil {
			log.Fatal(initErr)
		}
		defer initDb.Close()
		log.Println(fmt.Sprintf("Client database '%s' created in '%s'", sqliteFile, sqlitePath))

	} else {
		// a regular scanner processing event

		// coordinates for connecting to the sqlite database (from the command line options)
		dbCoordinates := database.ConnCoordinates{DBPath: sqlitePath, DBFile: sqliteFile}

		// attempt to connect to the sqlite db
		db, dbErr := database.InitializeDB(dbCoordinates)
		if dbErr != nil {
			log.Fatal(dbErr)
		}
		defer db.Close()

		processScanFn := func(barcode string) {
			// Lookup the barcode in the API server
			apiResponse, apiErr := http.PostForm(fmt.Sprintf("%s:%d/lookup", apiServer, apiPort), url.Values{"barcode": {barcode}})
			if apiErr != nil {
				fmt.Println(fmt.Sprintf("API access error: %s", apiErr))
				return
			}
			rawJson, _ := ioutil.ReadAll(apiResponse.Body)
			apiResponse.Body.Close()

			var products []*commerce.API
			err := json.Unmarshal(rawJson, &products)
			if err != nil {
				fmt.Println(fmt.Sprintf("API barcode lookup error: %s", err))
				return
			}

			// get the Account for this request
			acc, accErr := database.GetDesignatedAccount(db)
			if accErr != nil {
				fmt.Println(fmt.Sprintf("Client db account access error: %s", accErr))
				return
			}

			// get the list of current Vendors according to the Pi client database
			// and map them according to their API vendor id string
			vendors := make(map[string]*database.Vendor)
			for _, v := range database.GetAllVendors(db) {
				vendors[v.VendorId] = v
			}

			productsFound := 0
			for i, product := range products {
				v, exists := vendors[product.Vendor]
				if !exists {
					amazonId, amazonErr := database.AddVendor(db, product.Vendor, "Amazon")
					if amazonErr == nil {
						v = database.GetVendor(db, amazonId)
						vendors[product.Vendor] = v
					}
				}

				if len(product.ProductName) > 0 {
					// convert the commerce.API struct into a database.Item
					// so that it can be logged into the Pi client sqlite db
					item := database.Item{
						Index:           int64(i),
						Barcode:         barcode,
						Desc:            product.ProductName,
						UserContributed: false}
					pk, insertErr := item.Add(db, acc)
					if insertErr == nil {
						// also log the vendor/product code combination
						database.AddVendorProduct(db, product.SKU, v.Id, pk)
					}
					productsFound += 1
				}
			}

			if productsFound == 0 {
				// add it to the Pi client sqlite db as "unknown"
				// so that it can be manually edited/input
				unknownItem := database.Item{Index: 0, Barcode: barcode}
				unknownItem.Add(db, acc)
			}
		}

		errorFn := func(e error) {
			log.Fatal(e)
		}

		scanner.ScanForever(device, processScanFn, errorFn)
	}
}
