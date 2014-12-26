// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
	"github.com/Banrai/PiScan/client/database"
	"github.com/Banrai/PiScan/server/digest"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// InputUnknownItem handles the form for user contributions of unknown
// barcode scans: a GET presents the form, and a POST responds to the
// user-contributed input
func InputUnknownItem(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// get the Account for this request
	acc, accErr := database.GetDesignatedAccount(db)
	if accErr != nil {
		http.Error(w, accErr.Error(), http.StatusInternalServerError)
		return
	}

	// get the api server + port from the optional parameters
	apiHost, apiHostOk := opts[0].(string)
	if !apiHostOk {
		http.Error(w, BAD_REQUEST, http.StatusInternalServerError)
		return
	}

	// prepare the html page response
	form := &ItemForm{Title: "Contribute Product Information",
		CancelUrl:    HOME_URL,
		Unregistered: (acc.Email == database.ANONYMOUS_EMAIL)}

	//lookup the item from the request id
	// and show the input form (if a GET)
	// or process it (if a POST)
	if "GET" == r.Method {
		// derive the item id from the url path
		urlPaths := strings.Split(r.URL.Path[1:], "/")
		if len(urlPaths) >= 2 {
			itemId, itemIdErr := strconv.ParseInt(urlPaths[1], 10, 64)
			if itemIdErr == nil {
				item, itemErr := database.GetSingleItem(db, acc, itemId)
				if itemErr == nil {
					if item.Id != database.BAD_PK && item.Desc == "" {
						// requested item has been found and is valid
						form.Item = item
					}
				}
			}
		}

		if form.Item == nil {
			// no matching item was found
			http.Error(w, BAD_REQUEST, http.StatusInternalServerError)
			return
		}

	} else if "POST" == r.Method {
		// get the item id from the posted data
		r.ParseForm()
		idVal, idExists := r.PostForm["item"]
		barcodeVal, barcodeExists := r.PostForm["barcode"]
		prodNameVal, prodNameExists := r.PostForm["prodName"]
		if idExists && barcodeExists && prodNameExists {
			itemId, itemIdErr := strconv.ParseInt(idVal[0], 10, 64)
			if itemIdErr != nil {
				form.FormError = itemIdErr.Error()
			} else {
				item, itemErr := database.GetSingleItem(db, acc, itemId)
				if itemErr != nil {
					form.FormError = itemErr.Error()
				} else {
					// the hidden barcode value must match the retrieved item
					if item.Barcode == barcodeVal[0] {
						// update the item in the local client db
						item.Desc = prodNameVal[0]
						item.UserContributed = true
						item.Update(db)

						// also need to mark the contribution to POD in the server
						if acc.Email != database.ANONYMOUS_EMAIL {
							// get the form's prodDesc, brandName, brandUrl data
							prodDesc, prodDescExists := r.PostForm["prodDesc"]
							brandName, brandNameExists := r.PostForm["brandName"]
							brandUrl, brandUrlExists := r.PostForm["brandUrl"]

							// ping the server with the contribution data
							ping := func() {
								v := url.Values{}
								v.Set("email", acc.Email)
								v.Set("barcode", barcodeVal[0])
								v.Set("prodName", prodNameVal[0])
								if prodDescExists {
									v.Set("prodDesc", prodDesc[0])
								}
								if brandNameExists {
									v.Set("brandName", brandName[0])
								}
								if brandUrlExists {
									v.Set("brandUrl", brandUrl[0])
								}

								// use the account api code as the digest key
								hmac := digest.GenerateDigest(acc.APICode, v.Encode())
								v.Set("hmac", hmac)

								res, err := http.PostForm(strings.Join([]string{apiHost, "/contribute/"}, ""), v)
								if err == nil {
									res.Body.Close()
								}
							}

							go ping() // do not wait for the server to reply

						}

						// return success
						http.Redirect(w, r, HOME_URL, http.StatusFound)
						return
					} else {
						// bad form post: the hidden barcode value does not match the retrieved item
						form.FormError = BAD_POST
					}
				}
			}
		} else {
			// required form parameters are missing
			form.FormError = BAD_POST
		}
	}

	renderItemEditTemplate(w, form)
}
