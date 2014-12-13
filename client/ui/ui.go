// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Banrai/PiScan/client/database"
	"github.com/Banrai/PiScan/server/api"
	"github.com/Banrai/PiScan/server/digest"
	"github.com/mxk/go-sqlite/sqlite3"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	BAD_REQUEST = "Sorry, that is an invalid request"
	BAD_POST    = "Sorry, we cannot respond to that request. Please try again."
	HOME_URL    = "/scanned/"
	ACCOUNT_URL = "/account/"
)

var (
	TEMPLATE_LIST = func(templatesFolder string, templateFiles []string) []string {
		t := make([]string, 0)
		for _, f := range templateFiles {
			t = append(t, path.Join(templatesFolder, f))
		}
		return t
	}

	ITEM_LIST_TEMPLATE_FILES    = []string{"items.html", "head.html", "navigation_tabs.html", "actions.html", "modal.html", "scripts.html"}
	ITEM_EDIT_TEMPLATE_FILES    = []string{"define_item.html", "head.html", "scripts.html"}
	ACCOUNT_EDIT_TEMPLATE_FILES = []string{"account.html", "head.html", "navigation_tabs.html", "modal.html", "scripts.html"}

	ITEM_LIST_TEMPLATES    *template.Template
	ITEM_EDIT_TEMPLATES    *template.Template
	ACCOUNT_EDIT_TEMPLATES *template.Template

	TEMPLATES_INITIALIZED = false
)

// Use this to redirect one request to another target (string)
func Redirect(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// Respond to requests using HTML templates and the standard Content-Type (i.e., "text/html")
func MakeHTMLHandler(fn func(http.ResponseWriter, *http.Request, database.ConnCoordinates, ...interface{}), db database.ConnCoordinates, opts ...interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db, opts...)
	}
}

// Respond to requests that are not "text/html" Content-Types (e.g., for ajax calls)
func MakeHandler(fn func(*http.Request, database.ConnCoordinates, ...interface{}) string, db database.ConnCoordinates, mediaType string, opts ...interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mediaType))
		data := fn(r, db, opts...)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

// Standardize how the API Server string is generated
func defineApiServer(apiHost string) string {
	return fmt.Sprintf("http://%s", apiHost)
}

/* JSON response struct */
type AjaxAck struct {
	Message string `json:"msg"`
	Error   string `json:"err,omitempty"`
}

/* HTML template structs */
type ActiveTab struct {
	Scanned   bool
	Favorites bool
	Account   bool
	ShowTabs  bool
}

type Action struct {
	Icon   string
	Link   string
	Action string
}

type ItemsPage struct {
	Title     string
	ActiveTab *ActiveTab
	Actions   []*Action
	Items     []*database.Item
	Account   *database.Account
	Scanned   bool
}

type ItemForm struct {
	Title        string
	Item         *database.Item
	CancelUrl    string
	FormError    string
	FormMessage  string
	Unregistered bool
}

type AccountForm struct {
	Title        string
	ActiveTab    *ActiveTab
	Account      *database.Account
	CancelUrl    string
	FormError    string
	Unregistered bool
}

/* General db access functions */

// getItems returns a list of scanned or favorited products, and the correct
// corresponding options for the HTML page template
func getItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, favorites bool) {
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

	// define the appropriate fetch item function
	fetch := func(db *sqlite3.Conn, acc *database.Account) ([]*database.Item, error) {
		if favorites {
			return database.GetFavoriteItems(db, acc)
		} else {
			return database.GetItems(db, acc)
		}
	}

	// get all the desired items for this Account
	items := make([]*database.Item, 0)
	itemList, itemsErr := fetch(db, acc)
	if itemsErr != nil {
		http.Error(w, itemsErr.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range itemList {
		items = append(items, item)
	}

	// actions
	actions := make([]*Action, 0)
	// commerce options
	for _, vendor := range database.GetAllVendors(db) {
		actions = append(actions, &Action{Link: fmt.Sprintf("/buy%s/", vendor.VendorId), Icon: "fa fa-shopping-cart", Action: fmt.Sprintf("Buy from %s", vendor.DisplayName)})
	}
	if acc.Email != database.ANONYMOUS_EMAIL {
		actions = append(actions, &Action{Link: "/email/", Icon: "fa fa-envelope", Action: "Email to me"})
	}
	if favorites {
		actions = append(actions, &Action{Link: "/unfavorite/", Icon: "fa fa-star-o", Action: "Remove from favorites"})
	} else {
		actions = append(actions, &Action{Link: "/favorite/", Icon: "fa fa-star", Action: "Add to favorites"})
	}
	actions = append(actions, &Action{Link: "/delete/", Icon: "fa fa-trash", Action: "Delete"})

	// define the page title
	var titleBuffer bytes.Buffer
	if favorites {
		titleBuffer.WriteString("Favorite")
	} else {
		titleBuffer.WriteString("Scanned")
	}
	titleBuffer.WriteString(" Item")
	if len(itemList) != 1 {
		titleBuffer.WriteString("s")
	}

	p := &ItemsPage{Title: titleBuffer.String(),
		Scanned:   !favorites,
		ActiveTab: &ActiveTab{Scanned: !favorites, Favorites: favorites, Account: false, ShowTabs: true},
		Actions:   actions,
		Account:   acc,
		Items:     items}

	renderItemListTemplate(w, p)
}

// deleteItem attempts to lookup and remove the Item for the Account and
// Item.Id combination, returning a bool on success/fail, and the db lookup
// error (if any)
func deleteItem(db *sqlite3.Conn, acc *database.Account, id int64) (bool, error) {
	result := false

	item, itemErr := database.GetSingleItem(db, acc, id)
	if itemErr == nil {
		if item.Id == id {
			item.Delete(db)
			result = true
		}
	}

	return result, itemErr
}

// processItems fetches all the Items for the given Account, and the compares
// them to the id list posted from the form. All the matches get applied
// the given function: delete, favorite, unfavorite, etc.
func processItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, fn func(*database.Item, *sqlite3.Conn), successTarget string) {
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

	// get all the Items for this Account
	// and store them in a map by their Idscanned/
	items, itemsErr := database.GetItems(db, acc)
	if itemsErr != nil {
		http.Error(w, itemsErr.Error(), http.StatusInternalServerError)
		return
	}
	accountItems := make(map[int64]*database.Item)
	for _, item := range items {
		accountItems[item.Id] = item
	}

	// get the list of item ids from the POST values
	// and apply the processing function
	if "POST" == r.Method {
		r.ParseForm()
		if idVals, exists := r.PostForm["item"]; exists {
			for _, idString := range idVals {
				id, idErr := strconv.ParseInt(idString, 10, 64)
				if idErr == nil {
					if accountItem, ok := accountItems[id]; ok {
						fn(accountItem, db)
					}
				}
			}
		}
	}

	// finally, return home, to the scanned items list
	http.Redirect(w, r, successTarget, http.StatusFound)
}

/* HTML Response Functions (via templates) */

func renderItemListTemplate(w http.ResponseWriter, p *ItemsPage) {
	if TEMPLATES_INITIALIZED {
		ITEM_LIST_TEMPLATES.Execute(w, p)
	}
}

func renderItemEditTemplate(w http.ResponseWriter, f *ItemForm) {
	if TEMPLATES_INITIALIZED {
		ITEM_EDIT_TEMPLATES.Execute(w, f)
	}
}

func renderAccountEditTemplate(w http.ResponseWriter, a *AccountForm) {
	if TEMPLATES_INITIALIZED {
		ACCOUNT_EDIT_TEMPLATES.Execute(w, a)
	}
}

// InitializeTemplates confirms the given folder string leads to the html
// template files, otherwise templates.Must() will complain
func InitializeTemplates(folder string) {
	ITEM_LIST_TEMPLATES = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, ITEM_LIST_TEMPLATE_FILES)...))
	ITEM_EDIT_TEMPLATES = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, ITEM_EDIT_TEMPLATE_FILES)...))
	ACCOUNT_EDIT_TEMPLATES = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, ACCOUNT_EDIT_TEMPLATE_FILES)...))
	TEMPLATES_INITIALIZED = true
}

// ScannedItems returns all the products scanned, favorited or not, barcode
// lookup successful or not
func ScannedItems(w http.ResponseWriter, r *http.Request, db database.ConnCoordinates, opts ...interface{}) {
	getItems(w, r, db, false)
}

// FavoritedItems returns all the products scanned and favorited by this
// Account
func FavoritedItems(w http.ResponseWriter, r *http.Request, db database.ConnCoordinates, opts ...interface{}) {
	getItems(w, r, db, true)
}

// DeleteItems accepts a form post of one or more Item.Id values, and
// attempts to remove them from the client db. Unless it hits a critical
// error, it returns home, to the list of scanned items
func DeleteItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	del := func(i *database.Item, db *sqlite3.Conn) {
		i.Delete(db)
	}
	processItems(w, r, dbCoords, del, "/")
}

// FavoriteItems accepts a form post of one or more Item.Id values, and
// attempts to change their status in the client db to 'favorite'
func FavoriteItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	fav := func(i *database.Item, db *sqlite3.Conn) {
		i.Favorite(db)
	}
	processItems(w, r, dbCoords, fav, "/favorites/")
}

// UnfavoriteItems accepts a form post of one or more Item.Id values,
// and attempts to change their status in the client db to not 'favorite'
func UnfavoriteItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) {
	fav := func(i *database.Item, db *sqlite3.Conn) {
		i.Unfavorite(db)
	}
	processItems(w, r, dbCoords, fav, "/favorites/")
}

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

						// also need to mark the contribution in the POD clone db (via API, ultimately)
						// this is where to use the prodDesc, brandName, brandUrl post data

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

/* Ajax Response Functions (as strings via MakeHandler) */

// RemoveSingleItem looks up the single item represented by the itemId form
// post variable, and attempts to delete it, if it exists. The reply is a
// jsonified string, passed back to MakeHandler() to be coupled with the
// right mime type
func RemoveSingleItem(r *http.Request, dbCoords database.ConnCoordinates, opts ...interface{}) string {
	// prepare the ajax reply object
	ack := AjaxAck{Message: "", Error: ""}

	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		ack.Error = err.Error()
	}
	defer db.Close()

	if err == nil {
		// get the Account for this request
		acc, accErr := database.GetDesignatedAccount(db)
		if accErr != nil {
			ack.Error = accErr.Error()
		}

		// find the specific Item to remove
		// get the item id from the POST values
		if "POST" == r.Method {
			r.ParseForm()
			if idVal, exists := r.PostForm["itemId"]; exists {
				if len(idVal) > 0 {
					id, idErr := strconv.ParseInt(idVal[0], 10, 64)
					if idErr != nil {
						ack.Error = idErr.Error()
					} else {
						deleteSuccess, deleteErr := deleteItem(db, acc, id)
						if deleteSuccess {
							ack.Message = "Ok"
						} else {
							if deleteErr != nil {
								ack.Error = deleteErr.Error()
							} else {
								ack.Error = "No such item"
							}
						}
					}
				} else {
					ack.Error = "Missing item id"
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
