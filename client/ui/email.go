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

// EmailItems handles the client form post, to send a list of the selected
// items via email to the given user
func EmailItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// get the api server + port from the optional parameters
	apiHost, apiHostOk := opts[0].(string)
	if !apiHostOk {
		http.Error(w, BAD_REQUEST, http.StatusInternalServerError)
		return
	}

	// get the Account for this request
	acc, accErr := database.GetDesignatedAccount(db)
	if accErr != nil {
		http.Error(w, accErr.Error(), http.StatusInternalServerError)
		return
	}

	// get the account from the POST values
	if "POST" == r.Method {
		r.ParseForm()

		// make sure the form sent the required params
		accVal, accExists := r.PostForm["account"]
		items, itemsExist := r.PostForm["item"]

		if itemsExist && accExists {
			if len(accVal) > 0 {
				// check the submitted account info
				id, idErr := strconv.ParseInt(accVal[0], 10, 64)
				if idErr != nil {
					http.Error(w, idErr.Error(), http.StatusInternalServerError)
					return
				} else {
					if acc.Id != id {
						http.Error(w, BAD_REQUEST, http.StatusInternalServerError)
						return
					} else {
						// proceed with the send only if registered
						if acc.Email != database.ANONYMOUS_EMAIL {
							// lookup all the items for this account
							accountItems, accountItemsErr := database.GetItems(db, acc)

							if accountItemsErr == nil {
								// ping the server with the selected item data
								ping := func() {
									v := url.Values{}
									v.Set("email", acc.Email)

									// attach the list of descriptions for the selected items
									for _, accItem := range accountItems {
										for _, item := range items {
											if strconv.FormatInt(accItem.Id, 10) == item {
												v.Add("item", accItem.Desc)
												break
											}
										}
									}

									// use the account api code as the digest key
									hmac := digest.GenerateDigest(acc.APICode, v.Encode())
									v.Set("hmac", hmac)

									res, err := http.PostForm(strings.Join([]string{apiHost, "/email/"}, ""), v)
									if err == nil {
										res.Body.Close()
									}
								}

								go ping() // do not wait for the server to reply
							}
						}
					}
				}
			} else {
				http.Error(w, "Missing account id", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, BAD_POST, http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, BAD_REQUEST, http.StatusInternalServerError)
		return
	}

	// finally, return home, to the scanned items list
	http.Redirect(w, r, "/", http.StatusFound)

}
