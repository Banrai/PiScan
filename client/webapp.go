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

// have each request handler open its own db connection per request
func makeHandler(fn func(database.ConnCoordinates, http.ResponseWriter, *http.Request), db database.ConnCoordinates) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(db, w, r)
	}
}

func main() {
	var templatesFolder, dbPath, dbFile, dbTablesDefinitionPath string
	flag.StringVar(&templatesFolder, "templates", "", "Path to the html templates")
	flag.StringVar(&dbPath, "dbPath", "", "Path to the sqlite file")
	flag.StringVar(&dbFile, "dbFile", database.SQLITE_FILE, fmt.Sprintf("The sqlite database file (defaults to '%s')", database.SQLITE_FILE))
	flag.StringVar(&dbTablesDefinitionPath, "tables", "", fmt.Sprintf("Path to the sqlite database definitions file (%s)", database.TABLE_SQL_DEFINITIONS))
	flag.Parse()

	/* set the server ready for use */
	// confirm the html templates
	ui.InitializeTemplates(templatesFolder)

	// coordinates for connecting to the sqlite database
	dbCoordinates := database.ConnCoordinates{dbPath, dbFile, dbTablesDefinitionPath}

	/* define the server handlers */
	// dynamic request handlers
	http.HandleFunc("/", makeHandler(ui.ScannedItems, dbCoordinates))

	// static resources
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(templatesFolder, "../css/")))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(templatesFolder, "../js/")))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir(path.Join(templatesFolder, "../fonts/")))))

	/* start the server */
	http.ListenAndServe(":8080", nil)
}
