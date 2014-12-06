// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Banrai/PiScan/server/commerce"
	"github.com/Banrai/PiScan/server/commerce/amazon"
	"github.com/Banrai/PiScan/server/database/barcodes"
	"net/http"
	"strings"
)

// Lookup the barcode, using both the barcodes database, and the Amazon API
func LookupBarcode(r *http.Request, db DBConnection) string {
	// the result is a json representation of the list of found products
	products := make([]*commerce.API, 0)

	// this function only responds to POST requests
	if "POST" == r.Method {
		r.ParseForm()

		barcodeVal, barcodeExists := r.PostForm["barcode"]
		if barcodeExists {
			queryFn := func(statements map[string]*sql.Stmt) {
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
			WithServerDatabase(db, queryFn)
		}
	}

	result, err := json.Marshal(products)
	if err != nil {
		fmt.Println(err)
	}
	return string(result)
}
