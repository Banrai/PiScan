// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"database/sql"
	"fmt"
	"github.com/Banrai/PiScan/server/database/barcodes"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	mux    *http.ServeMux
	s      *http.Server
	Logger *log.Logger
}

type DBConnection struct {
	Host string
	User string
	Pass string
	Port int
}

var (
	Srv                      *Server
	DefaultServerReadTimeout = 30 // in seconds
)

func InitServerDatabase(dbCoords DBConnection) map[string]*sql.Stmt {
	statements := map[string]*sql.Stmt{}

	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/product_open_data", dbCoords.User, dbCoords.Pass, dbCoords.Host, dbCoords.Port)
	db, err := sql.Open("mysql", connection)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gtin, err := db.Prepare(barcodes.GTIN_LOOKUP)
	if err != nil {
		log.Fatal(err)
	} else {
		statements[barcodes.GTIN_LOOKUP] = gtin
	}
	defer gtin.Close()

	brand, err := db.Prepare(barcodes.BRAND_LOOKUP)
	if err != nil {
		log.Fatal(err)
	} else {
		statements[barcodes.BRAND_LOOKUP] = brand
	}
	defer brand.Close()

	asinLookup, err := db.Prepare(barcodes.ASIN_LOOKUP)
	if err != nil {
		log.Fatal(err)
	} else {
		statements[barcodes.ASIN_LOOKUP] = asinLookup
	}
	defer asinLookup.Close()

	asinInsert, err := db.Prepare(barcodes.ASIN_INSERT)
	if err != nil {
		log.Fatal(err)
	} else {
		statements[barcodes.ASIN_INSERT] = asinInsert
	}
	defer asinInsert.Close()

	return statements
}

func Respond(mediaType string, charset string, fn func(w http.ResponseWriter, r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=%s", mediaType, charset))
		data := fn(w, r)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

func NewAPIServer(host string, port int, timeout int, handlers map[string]func(http.ResponseWriter, *http.Request)) {
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.Handle(pattern, http.HandlerFunc(handler))
	}
	s := &http.Server{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Handler:     mux,
		ReadTimeout: time.Duration(timeout) * time.Second, // to prevent abuse of "keep-alive" requests by clients
	}
	Srv = &Server{
		mux:    mux,
		s:      s,
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
	Srv.s.ListenAndServe()
}
