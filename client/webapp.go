// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// This code initializes and runs the WebApp on the Pi client

package main

import (
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/client/database"
	"github.com/Banrai/PiScan/client/ui"
	"net/http"
	"path"
)

const (
	MIME_JSON = "application/json"
)

func main() {
	var templatesFolder, dbPath, dbFile string
	flag.StringVar(&templatesFolder, "templates", "", "Path to the html templates")
	flag.StringVar(&dbPath, "dbPath", database.SQLITE_PATH, fmt.Sprintf("Path to the sqlite file (defaults to '%s')", database.SQLITE_PATH))
	flag.StringVar(&dbFile, "dbFile", database.SQLITE_FILE, fmt.Sprintf("The sqlite database file (defaults to '%s')", database.SQLITE_FILE))
	flag.Parse()

	/* set the server ready for use */
	// confirm the html templates
	ui.InitializeTemplates(templatesFolder)

	// coordinates for connecting to the sqlite database (from the command line options)
	dbCoordinates := database.ConnCoordinates{DBPath: dbPath, DBFile: dbFile}

	/* define the server handlers */
	// dynamic request handlers
	http.HandleFunc("/", ui.Redirect("/scanned/"))
	http.HandleFunc("/scanned/", ui.MakeHTMLHandler(ui.ScannedItems, dbCoordinates))
	http.HandleFunc("/favorites/", ui.MakeHTMLHandler(ui.FavoritedItems, dbCoordinates))
	//http.HandleFunc("/buyAmazon/", makeHandler(ui.BuyFromAmazon, dbCoordinates))
	http.HandleFunc("/delete/", ui.MakeHTMLHandler(ui.DeleteItems, dbCoordinates))
	http.HandleFunc("/favorite/", ui.MakeHTMLHandler(ui.FavoriteItems, dbCoordinates))
	http.HandleFunc("/unfavorite/", ui.MakeHTMLHandler(ui.UnfavoriteItems, dbCoordinates))
	http.HandleFunc("/remove/", ui.MakeHandler(ui.RemoveSingleItem, dbCoordinates, MIME_JSON))
	//http.HandleFunc("/input/", makeHandler(ui.InputItem, dbCoordinates))

	// static resources
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(templatesFolder, "../css/")))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(templatesFolder, "../js/")))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir(path.Join(templatesFolder, "../fonts/")))))

	/* start the server */
	http.ListenAndServe(":8080", nil)
}
