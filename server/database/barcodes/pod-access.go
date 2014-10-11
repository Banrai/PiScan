// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package barcodes provides access to the database holding product data,
// sourced from both from the Open Product Database (POD) and every supported
// commerce API/site

package barcodes

import "database/sql"

const (
	// Prepared Queries (POD)
	GTIN_LOOKUP  = "select gtin_nm, bsin from gtin where gtin_cd = ?"
	BRAND_LOOKUP = "select brand_nm, brand_link from brand where bsin = ?"
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
