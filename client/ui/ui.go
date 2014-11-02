// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
	"bytes"
	"github.com/Banrai/PiScan/client/database"
	"github.com/mxk/go-sqlite/sqlite3"
	"html/template"
	"net/http"
	"path"
)

var (
	ITEM_TEMPLATE_FILES = []string{"base.html", "navigation_tabs.html", "actions.html", "items.html"}
	TEMPLATE_LIST       = func(templatesFolder string, templateFiles []string) []string {
		t := make([]string, 0)
		for _, f := range templateFiles {
			t = append(t, path.Join(templatesFolder, f))
		}
		return t
	}
	ITEM_TEMPLATES        *template.Template
	TEMPLATES_INITIALIZED = false
)

type ActiveTab struct {
	Scanned   bool
	Favorites bool
	ShowTabs  bool
}

type Action struct {
	Icon   string
	Link   string
	Action string
}

type Page struct {
	Title     string
	ActiveTab *ActiveTab
	Actions   []*Action
	Items     []*database.Item
	Scanned   bool
	ShowItems bool
}

func renderTemplate(w http.ResponseWriter, p *Page) {
	if TEMPLATES_INITIALIZED {
		ITEM_TEMPLATES.Execute(w, p)
	}
}

// getDesignatedAccount implements single-user mode (for now): it returns
// either the anonymous account, or the first non-anonymous account found
// on the sqlite database
func getDesignatedAccount(db *sqlite3.Conn) (*database.Account, error) {
	accounts, listErr := database.GetAllAccounts(db)
	if len(accounts) == 0 {
		return database.FetchAnonymousAccount(db)
	}
	return accounts[0], listErr
}

// InitializeTemplates confirms the given folder string leads to the html
// template files, otherwise templates.Must() will complain
func InitializeTemplates(folder string) {
	ITEM_TEMPLATES = template.Must(template.ParseFiles(TEMPLATE_LIST(folder, ITEM_TEMPLATE_FILES)...))
	TEMPLATES_INITIALIZED = true
}

// getItems returns a list of scanned or favorited products, and the correct
// correspoding options
func getItems(w http.ResponseWriter, r *http.Request, dbCoords database.ConnCoordinates, favorites bool) {
	// attempt to connect to the db
	db, err := database.InitializeDB(dbCoords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// get the Account for this request
	acc, accErr := getDesignatedAccount(db)
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
	itemList, itemsErr := fetch(db, acc) //database.GetItems(db, acc)
	if itemsErr != nil {
		http.Error(w, itemsErr.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range itemList {
		items = append(items, item)
	}

	// actions
	actions := make([]*Action, 0)
	actions = append(actions, &Action{Link: "/buyAmazon", Icon: "fa fa-shopping-cart", Action: "Buy from Amazon"})
	actions = append(actions, &Action{Link: "/email", Icon: "fa fa-envelope", Action: "Email to me"})
	actions = append(actions, &Action{Link: "/delete", Icon: "fa fa-trash", Action: "Delete"})
	if favorites {
		actions = append(actions, &Action{Link: "/unfavorite", Icon: "fa fa-star-o", Action: "Remove from favorites"})
	} else {
		actions = append(actions, &Action{Link: "/favorite", Icon: "fa fa-star", Action: "Add to favorites"})
	}

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

	p := &Page{Title: titleBuffer.String(),
		ShowItems: true,
		Scanned:   !favorites,
		ActiveTab: &ActiveTab{Scanned: !favorites, Favorites: favorites, ShowTabs: true},
		Actions:   actions,
		Items:     items}
	renderTemplate(w, p)
}

// ScannedItems returns all the products scanned, favorited or not, barcode
// lookup successful or not
func ScannedItems(dbCoords database.ConnCoordinates, w http.ResponseWriter, r *http.Request) {
	getItems(w, r, dbCoords, false)
}

// FavoritedItems returns all the products scanned and favorited by this
// Account
func FavoritedItems(dbCoords database.ConnCoordinates, w http.ResponseWriter, r *http.Request) {
	getItems(w, r, dbCoords, true)
}
