// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
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

func ScannedItems(dbCoords database.ConnCoordinates, w http.ResponseWriter, r *http.Request) {
	// Return a template view containing the specific list of scanned items

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

	// get all the scanned items for this Account
	items := make([]*database.Item, 0)
	itemList, itemsErr := database.GetItems(db, acc)
	if itemsErr != nil {
		http.Error(w, itemsErr.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range itemList {
		items = append(items, item)
	}

	// actions
	add := Action{Link: "/favorite", Icon: "fa fa-star-o", Action: "Add to favorites"}
	buy := Action{Link: "/buyAmazon", Icon: "fa fa-shopping-cart", Action: "Buy from Amazon"}

	p := &Page{Title: "PiScanner", ShowItems: true,
		ActiveTab: &ActiveTab{Scanned: true, Favorites: false, ShowTabs: true},
		Actions:   []*Action{&add, &buy},
		Items:     items}
	renderTemplate(w, p)
}
