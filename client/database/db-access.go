// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"github.com/mxk/go-sqlite/sqlite3"
	"io/ioutil"
	"path"
	"strings"
)

const (
	// Prepared Statements
	// User accounts
	GET_ACCOUNT  = "select id, api_code from account where email = $e"
	GET_ACCOUNTS = "select id, email, api_code from account"

	// Products
	GET_ITEMS          = "select id, barcode, product_desc, product_ind, datetime(posted) from product where account = $a"
	GET_FAVORITE_ITEMS = "select id, barcode, product_desc, product_ind, datetime(posted) from product where is_favorite = 1 and account = $a"
	DELETE_ITEM        = "delete from product where id = $i"
	FAVORITE_ITEM      = "update product set is_favorite = 1 where id = $i"
	UNFAVORITE_ITEM    = "update product set is_favorite = 0 where id = $i"
)

type Item struct {
	Id      int64
	Desc    string
	Barcode string
	Index   int
	Since   string
}

type Account struct {
	Id      int64
	Email   string
	APICode string
}

func InitializeDB(dbPath, dbName, tableSqlPath, tableSqlFile string) error {
	// attempt to open the sqlite db file at dbPath/dbName
	db, dbErr := sqlite3.Open(path.Join(dbPath, dbName))
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()

	// load the table definitions file from tableSqlPath/tableSqlFile
	// (tables.sql in this folder)
	content, err := ioutil.ReadFile(path.Join(tableSqlPath, tableSqlFile))
	if err != nil {
		return err
	}

	// attempt to create (if not exists) each table
	tables := strings.Split(string(content), ";")
	for _, table := range tables {
		db.Exec(table)
	}

	return nil
}
