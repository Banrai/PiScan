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

type SimpleMessage struct {
	Ack string
	Err error
}

var (
	Srv                      *Server
	DefaultServerReadTimeout = 30 // in seconds
)

func WithServerDatabase(dbCoords DBConnection, fn func(map[string]*sql.Stmt)) {
	preparedStatements := []string{barcodes.GTIN_LOOKUP,
		barcodes.BRAND_LOOKUP,
		barcodes.BRAND_NAME_LOOKUP,
		barcodes.BARCODE_LOOKUP,
		barcodes.BARCODE_INSERT,
		barcodes.BARCODE_BRAND_INSERT,
		barcodes.CONTRIBUTED_BRAND_LOOKUP,
		barcodes.CONTRIBUTED_BRAND_INSERT,
		barcodes.ASIN_LOOKUP,
		barcodes.ASIN_INSERT,
		barcodes.ACCOUNT_INSERT,
		barcodes.ACCOUNT_UPDATE,
		barcodes.ACCOUNT_DELETE,
		barcodes.ACCOUNT_LOOKUP_BY_EMAIL,
		barcodes.ACCOUNT_LOOKUP_BY_ID}

	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/product_open_data", dbCoords.User, dbCoords.Pass, dbCoords.Host, dbCoords.Port)
	db, err := sql.Open("mysql", connection)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statements := map[string]*sql.Stmt{}
	for _, p := range preparedStatements {
		stmt, err := db.Prepare(p)
		if err != nil {
			log.Fatal(err)
		} else {
			statements[p] = stmt
		}
	}

	fn(statements)
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
