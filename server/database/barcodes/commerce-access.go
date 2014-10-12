// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package barcodes provides access to the database holding product data,
// sourced from both from the Open Product Database (POD) and every supported
// commerce API/site

package barcodes

import "database/sql"

const (
	// Prepared Statements

	// Amazon
	ASIN_LOOKUP = "select asin, product, is_upc, is_ean, is_isbn, locale from amazon where barcode = ?"
	ASIN_INSERT = "insert into amazon (id, barcode, asin, product, is_upc, is_ean, is_isbn, locale) values (unhex(?), ?, ?, ?, ?, ?, ?, ?)"
)

// Data structures

// Amazon

type AMAZON struct {
	Asin        string `json:"asin"`
	Barcode     string `json:"barcode"`
	ProductName string `json:"product,omitempty"`
	ProductType string `json:"type,omitempty"`
	Locale      string `json:"locale"`
}

// Query Functions

// Amazon

// LookupAsin takes a prepared statment (using the ASIN_LOOKUP string), a
// barcode string, and looks it up in the Amazon table (sourced from querying
// the Amazon Product API), returning a list of matching AMAZON structs
func LookupAsin(stmt *sql.Stmt, barcode string) ([]*AMAZON, error) {
	results := make([]*AMAZON, 0)

	rows, err := stmt.Query(barcode)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			a, p, l        sql.NullString
			upc, ean, isbn sql.NullBool
		)

		err := rows.Scan(&a, &p, &upc, &ean, &isbn, &l)
		if err != nil {
			return results, err
		} else {
			if a.Valid && p.Valid {
				result := new(AMAZON)
				result.Barcode = barcode
				result.Asin = a.String
				result.ProductName = p.String
				result.Locale = l.String
				if upc.Valid && upc.Bool {
					result.ProductType = UPC
				} else if ean.Valid && ean.Bool {
					result.ProductType = EAN
				} else if isbn.Valid && isbn.Bool {
					result.ProductType = ISBN
				}
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// InsertAsin takes a prepared statement corresponding to an insert, an
// AMAZON data struct, and inserts the data from the struct as a new record
// in the database
func InsertAsin(stmt *sql.Stmt, rec AMAZON) error {
	isUpc  := false
	isEan  := false
	isIsbn := false

	if rec.ProductType == UPC {
		isUpc = true
	}
	if rec.ProductType == EAN {
		isEan = true
	}
	if rec.ProductType == ISBN {
		isIsbn = true
	}

	_, err := stmt.Exec(GenerateUUID(UndashedUUID), rec.Barcode, rec.Asin, rec.ProductName, isUpc, isEan, isIsbn, rec.Locale)

	return err
}
