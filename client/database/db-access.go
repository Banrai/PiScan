// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package database provides access to the sqlite database on the Pi client

package database

import (
	"fmt"
	"github.com/Banrai/PiScan/server/database/barcodes"
	"github.com/mxk/go-sqlite/sqlite3"
	"io/ioutil"
	"math"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	// Default database filename
	SQLITE_PATH = "/data"
	SQLITE_FILE = "PiScanDB.sqlite"

	// Default sql definitions file
	TABLE_SQL_DEFINITIONS = "tables.sql"

	// Execution constants
	BAD_PK = -1

	// Default Account (for those who don't want to register)
	ANONYMOUS_EMAIL = "anonymous@example.org"

	// Prepared Statements
	// User accounts
	ADD_ACCOUNT    = "insert into account (email, api_code) values ($e, $a)"
	GET_ACCOUNT    = "select id, api_code from account where email = $e"
	GET_ACCOUNTS   = "select id, email, api_code from account"
	UPDATE_ACCOUNT = "update account set email = $e, api_code = $a where id = $i"

	// Products
	ADD_ITEM           = "insert into product (barcode, product_desc, product_ind, is_edit, account) values ($b, $d, $i, $e, $a)"
	UPDATE_ITEM        = "update product set product_desc = $d, product_ind = $n, is_edit = $e where id = $i"
	GET_EXISTING_ITEM  = "select id from product where barcode = $b and product_desc = $d"
	GET_ITEMS          = "select id, barcode, product_desc, product_ind, strftime('%s', posted) from product where account = $a order by posted desc"
	GET_FAVORITE_ITEMS = "select id, barcode, product_desc, product_ind, strftime('%s', posted) from product where is_favorite = 1 and account = $a order by posted desc"
	DELETE_ITEM        = "delete from product where id = $i"
	FAVORITE_ITEM      = "update product set is_favorite = 1 where id = $i"
	UNFAVORITE_ITEM    = "update product set is_favorite = 0 where id = $i"

	// Commerce
	ADD_VENDOR         = "insert into vendor (vendor_id, display_name) values ($v, $n)"
	ADD_VENDOR_PRODUCT = "insert into product_availability (vendor, product_code, product) values ($v, $p, $i)"
	GET_VENDOR         = "select id, vendor_id, display_name from vendor where id = $i"
	GET_VENDORS        = "select distinct id, vendor_id, display_name from vendor"
	GET_VENDOR_PRODUCT = "select pa.id, v.id, pa.product_code from vendor v, product_availability pa where v.id = pa.vendor and pa.product = $i"
)

var (
	INTERVALS   = []string{"year", "month", "day", "hour", "minute"}
	SECONDS_PER = map[string]int64{"minute": 60, "hour": 3600, "day": 86400, "month": 2592000, "year": 31536000}
)

func calculateTimeSince(posted string) string {
	result := "just now" // default reply

	// try to convert the posted string into unix time
	i, err := strconv.ParseInt(posted, 10, 64)
	if err == nil {
		tm := time.Unix(i, 0)

		// calculate the time since posted
		// and return a human readable
		// '[interval] ago' string
		duration := time.Since(tm)
		if duration.Seconds() < 60.0 {
			if duration.Seconds() == 1.0 {
				result = fmt.Sprintf("%2.0f second ago", duration.Seconds())
			} else {
				result = fmt.Sprintf("%2.0f seconds ago", duration.Seconds())
			}
		} else {
			for _, interval := range INTERVALS {
				v := math.Trunc(duration.Seconds() / float64(SECONDS_PER[interval]))
				if v > 0.0 {
					if v == 1.0 {
						result = fmt.Sprintf("%2.0f %s ago", v, interval)
					} else {
						// plularize the interval label
						result = fmt.Sprintf("%2.0f %ss ago", v, interval)
					}
					break
				}
			}
		}
	}

	return result
}

func getPK(db *sqlite3.Conn, table string) int64 {
	// find and return the most recently-inserted
	// primary key, based on the table name
	sql := fmt.Sprintf("select seq from sqlite_sequence where name='%s'", table)

	var rowid int64
	for s, err := db.Query(sql); err == nil; err = s.Next() {
		s.Scan(&rowid)
	}
	return rowid
}

type ConnCoordinates struct {
	DBPath       string
	DBFile       string
	DBTablesPath string
}

type Account struct {
	Id      int64
	Email   string
	APICode string
}

type Vendor struct {
	Id          int64
	VendorId    string
	DisplayName string
}

type VendorProduct struct {
	Id          int64
	ProductCode string
	Vendor      *Vendor
}

type Item struct {
	Id              int64
	Desc            string
	Barcode         string
	Index           int64
	Since           string
	UserContributed bool
	ForSale         []*VendorProduct
}

func getExistingItem(db *sqlite3.Conn, barcode, desc string) int64 {
	// lookup the barcode and product desc
	// combination and return the primary key,
	// if the product has already been saved

	args := sqlite3.NamedArgs{"$b": barcode, "$d": desc}

	var rowid int64
	rowid = BAD_PK // default value, in case no match
	for s, err := db.Query(GET_EXISTING_ITEM, args); err == nil; err = s.Next() {
		s.Scan(&rowid)
	}
	return rowid
}

func (i *Item) Add(db *sqlite3.Conn, a *Account) (int64, error) {
	// insert the Item object

	// but first check if it's a duplicate or not
	itemPk := getExistingItem(db, i.Barcode, i.Desc)
	if itemPk != BAD_PK {
		return itemPk, nil
	}

	args := sqlite3.NamedArgs{"$b": i.Barcode,
		"$d": i.Desc,
		"$i": i.Index,
		"$e": i.UserContributed,
		"$a": a.Id}
	result := db.Exec(ADD_ITEM, args)
	if result == nil {
		pk := getPK(db, "product")
		return pk, result
	}

	return BAD_PK, result
}

func (i *Item) Update(db *sqlite3.Conn) error {
	// update the Item with with user contribution (description)
	args := sqlite3.NamedArgs{"$d": i.Desc,
		"$n": i.Index,
		"$e": i.UserContributed,
		"$i": i.Id}
	return db.Exec(UPDATE_ITEM, args)
}

func (i *Item) Delete(db *sqlite3.Conn) error {
	// delete the Item
	args := sqlite3.NamedArgs{"$i": i.Id}
	return db.Exec(DELETE_ITEM, args)
}

func (i *Item) Favorite(db *sqlite3.Conn) error {
	// update the Item, to show it is a favorite for this Account
	args := sqlite3.NamedArgs{"$i": i.Id}
	return db.Exec(FAVORITE_ITEM, args)
}

func (i *Item) Unfavorite(db *sqlite3.Conn) error {
	// update the Item, to show it is not a favorite for this Account
	args := sqlite3.NamedArgs{"$i": i.Id}
	return db.Exec(UNFAVORITE_ITEM, args)
}

func fetchItems(db *sqlite3.Conn, a *Account, sql string) ([]*Item, error) {
	// find all the items for this account
	results := make([]*Item, 0)

	args := sqlite3.NamedArgs{"$a": a.Id}
	row := make(sqlite3.RowMap)
	for s, err := db.Query(sql, args); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid, row)

		barcode, barcodeFound := row["barcode"]
		desc, descFound := row["product_desc"]
		ind, indFound := row["product_ind"]
		since, sinceFound := row["strftime('%s', posted)"]
		if barcodeFound {
			result := new(Item)
			result.Id = rowid
			result.Barcode = barcode.(string)
			if descFound {
				result.Desc = desc.(string)
			}
			if indFound {
				result.Index = ind.(int64)
			}
			if sinceFound {
				result.Since = calculateTimeSince(since.(string))
			}
			result.ForSale = GetVendorProducts(db, rowid)
			results = append(results, result)
		}
	}

	return results, nil
}

func GetItems(db *sqlite3.Conn, a *Account) ([]*Item, error) {
	return fetchItems(db, a, GET_ITEMS)
}

func GetFavoriteItems(db *sqlite3.Conn, a *Account) ([]*Item, error) {
	return fetchItems(db, a, GET_FAVORITE_ITEMS)
}

func GetSingleItem(db *sqlite3.Conn, a *Account, id int64) (*Item, error) {
	item := new(Item)
	item.Id = BAD_PK // if not found
	items, err := GetItems(db, a)
	for _, i := range items {
		if i.Id == id {
			return i, err
		}
	}
	return item, err
}

func AddVendor(db *sqlite3.Conn, vendorId, vendorDisplayName string) (int64, error) {
	args := sqlite3.NamedArgs{"$v": vendorId,
		"$n": vendorDisplayName}
	result := db.Exec(ADD_VENDOR, args)
	if result == nil {
		pk := getPK(db, "vendor")
		return pk, result
	}
	return BAD_PK, result
}

func AddVendorProduct(db *sqlite3.Conn, productCode string, vendorId, itemId int64) error {
	args := sqlite3.NamedArgs{"$v": vendorId,
		"$p": productCode,
		"$i": itemId}
	return db.Exec(ADD_VENDOR_PRODUCT, args)
}

func GetVendor(db *sqlite3.Conn, vendorId int64) *Vendor {
	result := new(Vendor)
	row := make(sqlite3.RowMap)
	args := sqlite3.NamedArgs{"$i": vendorId}
	for s, err := db.Query(GET_VENDOR, args); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid, row)
		vendor, vendorFound := row["vendor_id"]
		vendorName, vendorNameFound := row["display_name"]
		if vendorFound && vendorNameFound {
			result.Id = rowid
			result.VendorId = vendor.(string)
			result.DisplayName = vendorName.(string)
		}
	}
	return result
}

func GetAllVendors(db *sqlite3.Conn) []*Vendor {
	results := make([]*Vendor, 0)
	row := make(sqlite3.RowMap)
	for s, err := db.Query(GET_VENDORS); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid, row)
		vendor, vendorFound := row["vendor_id"]
		vendorName, vendorNameFound := row["display_name"]
		if vendorFound && vendorNameFound {
			v := Vendor{Id: rowid, VendorId: vendor.(string), DisplayName: vendorName.(string)}
			results = append(results, &v)
		}
	}
	return results
}

func GetVendorProducts(db *sqlite3.Conn, itemId int64) []*VendorProduct {
	results := make([]*VendorProduct, 0)

	row := make(sqlite3.RowMap)
	args := sqlite3.NamedArgs{"$i": itemId}
	for s, err := db.Query(GET_VENDOR_PRODUCT, args); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid, row)

		vendorPk, vendorPkFound := row["id"]
		productCode, productCodeFound := row["product_code"]
		if vendorPkFound && productCodeFound {
			result := new(VendorProduct)
			result.Id = rowid
			result.ProductCode = productCode.(string)
			result.Vendor = GetVendor(db, vendorPk.(int64))
			results = append(results, result)
		}
	}

	return results
}

func (a *Account) Add(db *sqlite3.Conn) error {
	// insert the Account object
	args := sqlite3.NamedArgs{"$e": a.Email, "$a": a.APICode}
	return db.Exec(ADD_ACCOUNT, args)
}

func (a *Account) Update(db *sqlite3.Conn, newEmail, newApi string) error {
	// update this Account's email and API code
	args := sqlite3.NamedArgs{"$i": a.Id, "$e": newEmail, "$a": newApi}
	return db.Exec(UPDATE_ACCOUNT, args)
}

func GetAccount(db *sqlite3.Conn, email string) (*Account, error) {
	// get the account corresponding to this email
	result := new(Account)

	args := sqlite3.NamedArgs{"$e": email}
	row := make(sqlite3.RowMap)
	for s, err := db.Query(GET_ACCOUNT, args); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid, row)

		api, apiFound := row["api_code"]
		if apiFound {
			result.APICode = api.(string)
			result.Id = rowid
			result.Email = email
			break
		}
	}

	return result, nil
}

func GetAllAccounts(db *sqlite3.Conn) ([]*Account, error) {
	// find all the accounts currently registered
	results := make([]*Account, 0)

	row := make(sqlite3.RowMap)
	for s, err := db.Query(GET_ACCOUNTS); err == nil; err = s.Next() {
		var rowid int64
		s.Scan(&rowid, row)

		email, emailFound := row["email"]
		api, apiFound := row["api_code"]
		if emailFound && apiFound {
			result := new(Account)
			result.APICode = api.(string)
			result.Id = rowid
			result.Email = email.(string)
			results = append(results, result)
		}
	}

	return results, nil
}

// FetchOrCreateDefaultAccount returns the existing local client account
// (in single-user mode), or creates it, if it does not exist yet
func FetchOrCreateDefaultAccount(db *sqlite3.Conn) (*Account, error) {
	// return the existing local client account
	anon, anonErr := GetAccount(db, ANONYMOUS_EMAIL)

	// or create it, if it does not exist yet
	if anon.Email == "" && anonErr == nil {
		anon = new(Account)
		anon.Email = ANONYMOUS_EMAIL
		anon.APICode = barcodes.GenerateUUID(barcodes.UndashedUUID)
		anonErr = anon.Add(db)
		if anonErr == nil {
			// make sure the Id value is correct
			return GetAccount(db, ANONYMOUS_EMAIL)
		}
	}

	return anon, anonErr
}

// GetDesignatedAccount implements single-user mode (for now): it returns
// either the anonymous account, or the first non-anonymous account found
// on the sqlite database
func GetDesignatedAccount(db *sqlite3.Conn) (*Account, error) {
	accounts, listErr := GetAllAccounts(db)
	if len(accounts) == 0 {
		return FetchOrCreateDefaultAccount(db)
	}
	return accounts[0], listErr
}

func InitializeDB(coords ConnCoordinates) (*sqlite3.Conn, error) {
	// attempt to open the sqlite db file
	db, dbErr := sqlite3.Open(path.Join(coords.DBPath, coords.DBFile))
	if dbErr != nil {
		return db, dbErr
	}

	// load the table definitions file, if coords.DBTablesPath is defined
	if len(coords.DBTablesPath) > 0 {
		content, err := ioutil.ReadFile(path.Join(coords.DBTablesPath, TABLE_SQL_DEFINITIONS))
		if err != nil {
			return db, err
		}

		// attempt to create (if not exists) each table
		tables := strings.Split(string(content), ";")
		for _, table := range tables {
			err = db.Exec(table)
			if err != nil {
				return db, err
			}
		}
	}

	return db, nil
}
