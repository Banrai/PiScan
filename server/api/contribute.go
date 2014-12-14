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
							//brandName, brandNameExists := r.PostForm["brandName"] // leave out the brand info for now
							//brandUrl, brandUrlExists := r.PostForm["brandUrl"]

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
								ack.Ack = fmt.Sprintf("ok: %s", pk)
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
