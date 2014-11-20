// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// This is a fully-functional (but simple) PiScanner application.

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/client/database"
	"github.com/Banrai/PiScan/scanner"
	"github.com/Banrai/PiScan/server/commerce/amazon"
	"github.com/Banrai/PiScan/server/database/barcodes"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

const ( // barcodes db lookups will be via internal api to a remote server, eventually
	barcodeDBUser   = "pod"
	barcodeDBServer = "127.0.0.1"
	barcodeDBPort   = "3306"
)

func main() {
	var device, dbUser, dbHost, dbPort, sqlitePath, sqliteFile, sqliteTablesDefinitionPath string
	flag.StringVar(&device, "device", scanner.SCANNER_DEVICE, fmt.Sprintf("The '/dev/input/event' device associated with your scanner (defaults to '%s')", scanner.SCANNER_DEVICE))
	flag.StringVar(&dbUser, "dbUser", barcodeDBUser, fmt.Sprintf("The barcodes database user (defaults to '%s')", barcodeDBUser))
	flag.StringVar(&dbHost, "dbHost", barcodeDBServer, fmt.Sprintf("The barcodes database server (defaults to '%s')", barcodeDBServer))
	flag.StringVar(&dbPort, "dbPort", barcodeDBPort, fmt.Sprintf("The barcodes database port (defaults to '%s')", barcodeDBPort))
	flag.StringVar(&sqlitePath, "sqlitePath", database.SQLITE_PATH, fmt.Sprintf("Path to the sqlite file (defaults to '%s')", database.SQLITE_PATH))
	flag.StringVar(&sqliteFile, "sqliteFile", database.SQLITE_FILE, fmt.Sprintf("The sqlite database file (defaults to '%s')", database.SQLITE_FILE))
	flag.StringVar(&sqliteTablesDefinitionPath, "sqliteTables", "", fmt.Sprintf("Path to the sqlite database definitions file, %s, (use only if creating the client db for the first time)", database.TABLE_SQL_DEFINITIONS))
	flag.Parse()

	if len(sqliteTablesDefinitionPath) > 0 {
		// this is a request to create the client db for the first time
		_, initErr := database.InitializeDB(database.ConnCoordinates{sqlitePath, sqliteFile, sqliteTablesDefinitionPath})
		if initErr != nil {
			log.Fatal(initErr)
		}
		log.Println(fmt.Sprintf("Client database '%s' created in '%s'", sqliteFile, sqlitePath))

	} else {
		// a regular scanner processing event

		// eventually, make all this via an internal api call to a remote server, but for now...
		// connect to the barcodes database and make all the prepared statements available to scanner functions
		db, err := sql.Open("mysql",
			strings.Join([]string{dbUser, "@tcp(", dbHost, ":", dbPort, ")/product_open_data"}, ""))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		gtin, err := db.Prepare(barcodes.GTIN_LOOKUP)
		if err != nil {
			log.Fatal(err)
		}
		defer gtin.Close()

		brand, err := db.Prepare(barcodes.BRAND_LOOKUP)
		if err != nil {
			log.Fatal(err)
		}
		defer brand.Close()

		asinLookup, err := db.Prepare(barcodes.ASIN_LOOKUP)
		if err != nil {
			log.Fatal(err)
		}
		defer asinLookup.Close()

		asinInsert, err := db.Prepare(barcodes.ASIN_INSERT)
		if err != nil {
			log.Fatal(err)
		}
		defer asinInsert.Close()

		// coordinates for connecting to the sqlite database (from the command line options)
		dbCoordinates := database.ConnCoordinates{DBPath: sqlitePath, DBFile: sqliteFile}

		processScanFn := func(barcode string) {
			// Lookup the barcode, using both the barcodes database, and the Amazon API
			products, err := amazon.Lookup(barcode, asinLookup, asinInsert)
			if err != nil {
				fmt.Println(fmt.Sprintf("Barcode lookup error: %s", err))
			} else {
				// attempt to connect to the sqlite db
				db, dbErr := database.InitializeDB(dbCoordinates)
				if err != nil {
					fmt.Println(fmt.Sprintf("Client db access error: %s", dbErr))
					return
				}
				defer db.Close()

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
					//fmt.Println(fmt.Sprintf("(%d) SKU %s Name %s Type %s Vendor %s", i, product.SKU, product.ProductName, product.ProductType, product.Vendor))
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
		}
		errorFn := func(e error) {
			log.Fatal(e)
		}
		scanner.ScanForever(device, processScanFn, errorFn)
	}
}
