// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
	"encoding/json"
	"github.com/Banrai/PiScan/client/database"
	"github.com/Banrai/PiScan/server/api"
	"github.com/Banrai/PiScan/server/digest"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	ACCOUNT_EDIT_TEMPLATE_FILES = []string{"account.html", "head.html", "navigation_tabs.html", "modal.html", "scripts.html"}
	ACCOUNT_EDIT_TEMPLATES      *template.Template
)

type AccountForm struct {
	Title        string
	ActiveTab    *ActiveTab
	Account      *database.Account
	CancelUrl    string
	FormError    string
	Unregistered bool
}

/* HTML Response Functions (via templates) */

func renderAccountEditTemplate(w http.ResponseWriter, a *AccountForm) {
	if TEMPLATES_INITIALIZED {
		ACCOUNT_EDIT_TEMPLATES.Execute(w, a)
	}
}

// EditAccount presents the form for editing Account information (in
// response to a GET request) and handles to add/updates (in response to
// a POST request)
func EditAccount(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
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
	regStatus := (acc.Email == database.ANONYMOUS_EMAIL)
	cancelUrl := HOME_URL
	if !regStatus {
		cancelUrl = ACCOUNT_URL
	}

	form := &AccountForm{Title: "My Account",
		ActiveTab:    &ActiveTab{Scanned: false, Favorites: false, Account: true, ShowTabs: true},
		Account:      acc,
		CancelUrl:    cancelUrl,
		Unregistered: regStatus}

	if "POST" == r.Method {
		form.FormError = BAD_POST // in event of problems

		// get the item id from the posted data
		r.ParseForm()
		accVal, accExists := r.PostForm["account"]
		emailVal, emailExists := r.PostForm["accountEmail"]
		if accExists && emailExists {
			// make sure the hidden account id value matches the Account
			accId, accIdErr := strconv.ParseInt(accVal[0], 10, 64)
			if accIdErr != nil {
				form.FormError = accIdErr.Error()
			} else {
				if acc.Id == accId {
					// update the account email address in the local client db
					updateErr := acc.Update(db, emailVal[0], acc.APICode)
					if updateErr != nil {
						form.FormError = updateErr.Error()
					} else {
						// ping the server with the api code and email for verification
						ping := func() {
							v := url.Values{}
							v.Set("email", emailVal[0])
							v.Set("api", acc.APICode)

							// use the email address as the digest key
							hmac := digest.GenerateDigest(emailVal[0], v.Encode())
							v.Set("hmac", hmac)

							res, err := http.Get(strings.Join([]string{defineApiServer(apiHost), "/register?", v.Encode()}, ""))
							if err == nil {
								res.Body.Close()
							}
						}

						go ping() // do not wait for the server to reply

						// return success
						http.Redirect(w, r, ACCOUNT_URL, http.StatusFound)
						return
					}
				}
			}
		}
	}

	renderAccountEditTemplate(w, form)
}

// ConfirmServerAccount responds to the ajax request from the client to
// lookup and return the status of the given account
func ConfirmServerAccount(r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) string {
	// prepare the ajax reply object
	ack := AjaxAck{Message: "", Error: ""}

	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		ack.Error = err.Error()
	}
	defer db.Close()

	// get the api server + port from the optional parameters
	apiHost, apiHostOk := opts[0].(string)
	if !apiHostOk {
		ack.Error = BAD_REQUEST
	}

	if ack.Error == "" {
		// get the Account for this request
		acc, accErr := database.GetDesignatedAccount(db)
		if accErr != nil {
			ack.Error = accErr.Error()
		}

		// get the account from the POST values
		if "POST" == r.Method {
			r.ParseForm()
			if accVal, exists := r.PostForm["account"]; exists {
				if len(accVal) > 0 {
					id, idErr := strconv.ParseInt(accVal[0], 10, 64)
					if idErr != nil {
						ack.Error = idErr.Error()
					} else {
						if acc.Id != id {
							ack.Error = BAD_REQUEST
						} else {
							// prepare the API Server request
							v := url.Values{}
							v.Set("email", acc.Email)

							// use the email address as the digest key
							hmac := digest.GenerateDigest(acc.Email, v.Encode())
							v.Set("hmac", hmac)

							// ping the API Server for the status of this account
							res, resErr := http.Get(strings.Join([]string{defineApiServer(apiHost), "/status?", v.Encode()}, ""))
							defer res.Body.Close()
							if resErr != nil {
								ack.Error = resErr.Error()
							} else {
								// read and parse the json message from the API Server
								m := new(api.SimpleMessage)
								dec := json.NewDecoder(res.Body)
								dec.Decode(&m)

								// assign the json ack accordingly
								if m.Err != nil {
									ack.Error = m.Err.Error()
								} else {
									ack.Message = m.Ack
								}
							}
						}
					}
				} else {
					ack.Error = "Missing account id"
				}
			} else {
				ack.Error = BAD_POST
			}
		} else {
			ack.Error = BAD_REQUEST
		}
	}

	// convert the ajax reply object to json
	ackObj, ackObjErr := json.Marshal(ack)
	if ackObjErr != nil {
		return ackObjErr.Error()
	}
	return string(ackObj)
}
