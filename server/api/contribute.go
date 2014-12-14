// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Banrai/PiScan/server/database/barcodes"
	"github.com/Banrai/PiScan/server/digest"
	"net/http"
)

func ContributeData(r *http.Request, db DBConnection) string {
	// the result is a simple json ack
	ack := new(SimpleMessage)

	// this function only responds to POST requests
	if "POST" == r.Method {
		r.ParseForm()

		emailVal, emailValExists := r.PostForm["email"]
		barcode, barcodeExists := r.PostForm["barcode"]
		hmacDigest, hmacDigestExists := r.PostForm["hmac"]

		if emailValExists && barcodeExists && hmacDigestExists {
			processFn := func(statements map[string]*sql.Stmt) {
				// see if the account exists
				accountLookupStmt, accountLookupStmtExists := statements[barcodes.ACCOUNT_LOOKUP_BY_EMAIL]
				itemInsertStmt, itemInsertStmtExists := statements[barcodes.BARCODE_INSERT]
				if accountLookupStmtExists && itemInsertStmtExists {
					// see if the email is available
					acc, accErr := barcodes.LookupAccount(accountLookupStmt, emailVal[0], false)
					if accErr != nil {
						ack.Err = accErr
					} else {
						// check the hmac digest
						r.PostForm.Del("hmac") // separate the digest from the rest
						if digest.DigestMatches(acc.APICode, r.PostForm.Encode(), hmacDigest[0]) {
							// hmac is correct

							// add the contributed barcode data
							prodName, prodNameExists := r.PostForm["prodName"]
							prodDesc, prodDescExists := r.PostForm["prodDesc"]

							item := barcodes.BARCODE{Barcode: barcode[0], GtinEdit: false}
							if prodNameExists {
								item.ProductName = prodName[0]
							}
							if prodDescExists {
								item.ProductDesc = prodDesc[0]
							}
							pk, insertErr := barcodes.ContributeBarcode(itemInsertStmt, item, acc)
							if insertErr != nil {
								ack.Err = insertErr
							} else {
								item.Uuid = pk
								ack.Ack = fmt.Sprintf("ok: %s", pk)

								// contribute the brand information, if any
								brandName, brandNameExists := r.PostForm["brandName"]
								brandUrl, brandUrlExists := r.PostForm["brandUrl"]
								if brandNameExists || brandUrlExists {
									// TO-DO: use an autocomplete/autosuggestion at the UI form,
									// so that what gets posted here is either definitely an existing BSIN or not ...
									// but instead, for now, use the name to lookup possible matches
									brandLookupStmt, brandLookupStmtExists := statements[barcodes.BRAND_NAME_LOOKUP]
									brandInsertStmt, brandInsertStmtExists := statements[barcodes.CONTRIBUTED_BRAND_INSERT]
									brandSupplementStmt, brandSuplementStmtExists := statements[barcodes.BARCODE_BRAND_INSERT]
									if brandLookupStmtExists && brandInsertStmtExists && brandSuplementStmtExists {
										// see if the brand already exists in POD
										existingBrands, existingBrandsErr := barcodes.LookupBrandByName(brandLookupStmt, brandName[0])
										if existingBrandsErr != nil {
											ack.Err = existingBrandsErr
										}
										if len(existingBrands) > 0 {
											// for now, just take the first match, and use it as the existing POD brand
											// mark the contributed barcode item as belonging to this brand
											ack.Err = barcodes.ContributeBarcodeBrand(brandSupplementStmt, item, existingBrands[0])
										} else {
											// this brand is completely unknown to POD
											brand := new(barcodes.CONTRIBUTED_BRAND)
											if brandNameExists {
												brand.Name = brandName[0]
											}
											if brandUrlExists {
												brand.URL = brandUrl[0]
											}
											ack.Err = barcodes.ContributeBrand(brandInsertStmt, brand, acc)
										}
									}
								}
							}
						}
					}
				}
			}
			WithServerDatabase(db, processFn)

		}
	}

	result, err := json.Marshal(ack)
	if err != nil {
		fmt.Println(err)
	}
	return string(result)
}
