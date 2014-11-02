// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package amazon provides methods for looking up barcodes and finding
// their associated Amazon catalog product information, either by using
// the Product API, or, if the particuar barcode has been found before,
// from the barcodes database

package amazon

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/Banrai/PiScan/server/commerce"
	"github.com/Banrai/PiScan/server/database/barcodes"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// The apiLookup function provides a simple interface to the python
// amazon_api_lookup.py script using os/exec and returns the string result
// and error (if any) as-is
func apiLookup(barcode string) (string, error) {

	// find the path of the calling binary
	_, filename, _, _ := runtime.Caller(1)

	// for now, the path to the API lookup script is relative to the source root
	// so pass the barcode string to it as the first command line argument and
	// capture and return the result
	lookupCmd := []string{"python", path.Join(path.Dir(filename), "/amazon_api_lookup.py"), barcode}
	cmd := exec.Command(lookupCmd[0], lookupCmd[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	return out.String(), err
}

// The Lookup function first looks for the given barcode in the barcodes
// database. If not found there, it tries the Amazon Product API, and save
// all those results into the barcodes database for future reference. It
// returns the json (a list of API structs, one per product) and error.
func Lookup(barcode string, asinLookup, asinInsert *sql.Stmt) ([]*commerce.API, error) {
	results := make([]*commerce.API, 0)
	var resultErr error

	// see if the barcode already exists in the db
	products, err := barcodes.LookupAsin(asinLookup, barcode)
	resultErr = err
	if err == nil && len(products) > 0 {
		for _, product := range products {
			result := new(commerce.API)
			result.SKU = product.Asin
			result.ProductName = product.ProductName
			result.ProductType = product.ProductType
			result.Vendor = strings.Join([]string{"AMZN", product.Locale}, ":")
			results = append(results, result)
		}
	} else {
		// if not, use the API instead, and save any results to the barcodes db
		api, aerr := apiLookup(barcode)
		resultErr = aerr
		if aerr == nil {
			// convert the api result string into a json object
			var apiList []commerce.API
			jerr := json.Unmarshal([]byte(api), &apiList)
			if jerr == nil {
				for _, apiResult := range apiList {
					// save each result for re-marshalling into json
					results = append(results, &apiResult)

					// and save it in the db, for the future
					prod := new(barcodes.AMAZON)
					prod.Barcode = barcode
					prod.Asin = apiResult.SKU
					prod.ProductName = apiResult.ProductName
					prod.ProductType = apiResult.ProductType
					vendor := strings.Split(apiResult.Vendor, ":")
					if len(vendor) > 0 {
						prod.Locale = vendor[1]
					} else {
						// use the default
						prod.Locale = "us"
					}
					_ = barcodes.InsertAsin(asinInsert, *prod)
				}
			}
		}
	}

	return results, resultErr
}
