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

/*
type API struct {
	SKU         string `json:"sku"`
	ProductName string `json:"desc,omitempty"`
	ProductType string `json:"type,omitempty"`
	Vendor      string `json:"vnd"`
}

type GTIN struct {
	Id          string `json:"barcode"`
	ProductName string `json:"product,omitempty"`
	BrandId     string `json:"bsin,omitempty"`
}

type BARCODE struct {
	Uuid        string `json:"id"`
	Barcode     string `json:"barcode"`
	ProductName string `json:"product,omitempty"`
	ProductDesc string `json:"desc,omitempty"`
	GtinEdit    bool   `json:"gtinCorrection,omitempty"`
	AccountID   string `json:"account"`
}
*/

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
				barcode := strings.Join(barcodeVal, "")

				// lookup the barcode versus the regular POD db
				podLookup, podLookupExists := statements[barcodes.GTIN_LOOKUP]
				if podLookupExists {
					podMatches, podMatchErr := barcodes.LookupGtin(podLookup, barcode)
					if podMatchErr == nil {
						for _, podMatch := range podMatches {
							// convert each podMatch GTIN struct to a commerce.API struct
							m := new(commerce.API)
							m.SKU = barcode
							m.ProductName = podMatch.ProductName
							products = append(products, m)
						}
					}
				}

				// lookup the barcode versus the Amazon db table/API
				asinLookup, asinLookupExists := statements[barcodes.ASIN_LOOKUP]
				asinInsert, asinInsertExists := statements[barcodes.ASIN_INSERT]
				if asinLookupExists && asinInsertExists {
					prods, prodErr := amazon.Lookup(barcode, asinLookup, asinInsert)
					if prodErr == nil {
						for _, prod := range prods {
							products = append(products, prod)
						}
					}
				}

				// supplement the list of results by looking at the user contributions
				contribLookup, contribLookupExists := statements[barcodes.BARCODE_LOOKUP]
				if contribLookupExists {
					contribMatches, contribMatchErr := barcodes.LookupContributedBarcode(contribLookup, barcode)
					if contribMatchErr == nil {
						for _, contrib := range contribMatches {
							// convert each contribMatch BARCODE struct to a commerce.API struct
							c := new(commerce.API)
							c.SKU = barcode
							c.ProductName = contrib.ProductName
							if contrib.ProductDesc != "" {
								c.ProductType = contrib.ProductDesc
							}
							products = append(products, c)
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
