// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package barcodes provides access to the database holding product data,
// sourced from both from the Open Product Database (POD) and every supported
// commerce API/site

package barcodes

import (
	"database/sql"
	"fmt"
)

const (
	// Prepared Queries (POD)
	GTIN_LOOKUP       = "select gtin_nm, bsin from gtin where gtin_cd = ?"
	BRAND_LOOKUP      = "select brand_nm, brand_link from brand where bsin = ?"
	BRAND_NAME_LOOKUP = "select bsin, brand_nm, brand_link from brand where brand_nm like ?"

	// User contributions
	BARCODE_LOOKUP           = "select hex(id), product_name, product_desc, is_edit, hex(account_id) from barcode where barcode = ?"
	BARCODE_INSERT           = "insert into barcode (id, barcode, product_name, product_desc, is_edit, account_id) values (unhex(?), ?, ?, ?, ?, unhex(?))"
	BARCODE_BRAND_INSERT     = "insert into barcode_brand (id, bsin, barcode_id) values (unhex(?), ?, unhex(?))"
	CONTRIBUTED_BRAND_LOOKUP = "select hex(id), brand_name, brand_url, hex(account_id) from contributed_brand where brand_name like ?"
	CONTRIBUTED_BRAND_INSERT = "insert into contributed_brand (id, brand_name, brand_url, account_id) values (unhex(?), ?, ?, unhex(?))"
)

// Data structures (POD)

type GTIN struct {
	Id          string `json:"barcode"`
	ProductName string `json:"product,omitempty"`
	BrandId     string `json:"bsin,omitempty"`
}

type BRAND struct {
	Id   string `json:"bsin"`
	Name string `json:"brand,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Data structures (user contributions)

type BARCODE struct {
	Uuid        string `json:"id"`
	Barcode     string `json:"barcode"`
	ProductName string `json:"product,omitempty"`
	ProductDesc string `json:"desc,omitempty"`
	GtinEdit    bool   `json:"gtinCorrection,omitempty"`
	AccountID   string `json:"account"`
}

type CONTRIBUTED_BRAND struct {
	Uuid      string `json:"id"`
	Name      string `json:"brand,omitempty"`
	URL       string `json:"url,omitempty"`
	AccountID string `json:"account"`
}

// Query Functions (POD)

// LookupGtin takes a prepared statment (using the GTIN_LOOKUP string),
// a barcode string, and looks it up in POD, returning a list of matching
// GTIN structs
func LookupGtin(stmt *sql.Stmt, barcode string) ([]*GTIN, error) {
	results := make([]*GTIN, 0)

	rows, err := stmt.Query(barcode)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var p, b sql.NullString
		err := rows.Scan(&p, &b)
		if err != nil {
			return results, err
		} else {
			if p.Valid || b.Valid {
				result := new(GTIN)
				result.Id = barcode
				if p.Valid {
					result.ProductName = p.String
				}
				if b.Valid {
					result.BrandId = b.String
				}
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// LookupBrand takes a prepared statment (using the BRAND_LOOKUP string),
// a bsin (brand id) string, and looks it up in POD, returning a list of
// matching BRAND structs
func LookupBrand(stmt *sql.Stmt, bsin string) ([]*BRAND, error) {
	results := make([]*BRAND, 0)

	rows, err := stmt.Query(bsin)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var n, u sql.NullString
		err := rows.Scan(&n, &u)
		if err != nil {
			return results, err
		} else {
			if n.Valid || u.Valid {
				result := new(BRAND)
				result.Id = bsin
				if n.Valid {
					result.Name = n.String
				}
				if u.Valid {
					result.URL = u.String
				}
				results = append(results, result)
			}
		}
	}

	return results, nil
}

func LookupBrandByName(stmt *sql.Stmt, brandName string) ([]*BRAND, error) {
	results := make([]*BRAND, 0)

	rows, err := stmt.Query(fmt.Sprintf("%s%%", brandName)) // like 'brandName%'
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var i, n, u sql.NullString
		err := rows.Scan(&i, &n, &u)
		if err != nil {
			return results, err
		} else {
			if n.Valid || u.Valid {
				result := new(BRAND)
				result.Id = i.String
				if n.Valid {
					result.Name = n.String
				}
				if u.Valid {
					result.URL = u.String
				}
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// Query Functions (user contributions)

func LookupContributedBarcode(stmt *sql.Stmt, code string) ([]*BARCODE, error) {
	results := make([]*BARCODE, 0)

	rows, err := stmt.Query(code)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			i, n, d, a sql.NullString
			e          sql.NullBool
		)
		err := rows.Scan(&i, &n, &d, &e, &a)
		if err != nil {
			return results, err
		} else {
			result := new(BARCODE)
			result.Barcode = code
			result.Uuid = i.String
			result.ProductName = n.String
			result.GtinEdit = e.Bool
			result.AccountID = a.String
			if d.Valid {
				result.ProductDesc = d.String
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func LookupContributedBrand(stmt *sql.Stmt, brandName string) ([]*CONTRIBUTED_BRAND, error) {
	results := make([]*CONTRIBUTED_BRAND, 0)

	rows, err := stmt.Query(fmt.Sprintf("%s%%", brandName)) // like 'brandName%'
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var i, n, u, a sql.NullString
		err := rows.Scan(&i, &n, &u, &a)
		if err != nil {
			return results, err
		} else {
			if n.Valid {
				result := new(CONTRIBUTED_BRAND)
				result.Uuid = i.String
				result.Name = n.String
				result.URL = u.String
				result.AccountID = a.String
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// Write functions (user contributions)

func ContributeBarcode(stmt *sql.Stmt, rec BARCODE, acc *ACCOUNT) (string, error) {
	if rec.Uuid == "" {
		rec.Uuid = GenerateUUID(UndashedUUID)
	}

	_, err := stmt.Exec(rec.Uuid, rec.Barcode, rec.ProductName, rec.ProductDesc, rec.GtinEdit, acc.Id)

	return rec.Uuid, err
}

func ContributeBarcodeBrand(stmt *sql.Stmt, rec BARCODE, brand *BRAND) error {
	_, err := stmt.Exec(GenerateUUID(UndashedUUID), brand.Id, rec.Uuid)

	return err
}

func ContributeBrand(stmt *sql.Stmt, rec *CONTRIBUTED_BRAND, acc *ACCOUNT) error {
	_, err := stmt.Exec(GenerateUUID(UndashedUUID), rec.Name, rec.URL, acc.Id)

	return err
}
