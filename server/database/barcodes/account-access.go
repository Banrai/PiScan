// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package barcodes provides access to the database holding product data,
// sourced from both from the Open Product Database (POD) and every supported
// commerce API/site

package barcodes

import "database/sql"

const (
	// Prepared Queries (POD contributor accounts)

	// Create/Update/Delete
	ACCOUNT_INSERT = "insert into account (id, email, verify_code) values (unhex(?), ?, ?)"
	ACCOUNT_UPDATE = "update account set email = ?, verified = ?, enabled = ?, date_verified=NOW() where id = unhex(?)"
	ACCOUNT_DELETE = "delete from account where id = unhex(?)"

	// Lookup
	ACCOUNT_LOOKUP_BY_EMAIL = "select hex(id), verify_code, verified, enabled from account where email = ?"
	ACCOUNT_LOOKUP_BY_ID    = "select email, verify_code, verified, enabled from account where id = unhex(?)"
)

// Data structure (POD contributor accounts)

type ACCOUNT struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	APICode  string `json:"code"`
	Verified bool   `json:"verified,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
}

// Query Functions

// LookupAccount searches for the Account using either the id (uuid) or
// email address string parameter, depending on which prepared statement
// and boolean flag is passed
func LookupAccount(stmt *sql.Stmt, param string, usingId bool) (*ACCOUNT, error) {
	result := new(ACCOUNT)

	rows, err := stmt.Query(param)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			n, cd sql.NullString
			v, e  sql.NullBool
		)

		err := rows.Scan(&n, &cd, &v, &e)
		if err != nil {
			return result, err
		} else {
			if usingId {
				result.Id = param
				result.Email = n.String
			} else {
				result.Id = n.String
				result.Email = param
			}
			result.APICode = cd.String
			result.Verified = v.Bool
			result.Enabled = e.Bool

			break
		}
	}

	return result, nil
}

// Add, Update, and Delete

func (a *ACCOUNT) Add(stmt *sql.Stmt) (string, error) {
	pk := GenerateUUID(UndashedUUID)
	_, err := stmt.Exec(pk, a.Email, a.APICode)

	return pk, err
}

func (a *ACCOUNT) Delete(stmt *sql.Stmt) error {
	_, err := stmt.Exec(a.Id)

	return err
}

func (a *ACCOUNT) Update(stmt *sql.Stmt) error {
	_, err := stmt.Exec(a.Email, a.APICode, a.Verified, a.Enabled, a.Id)

	return err
}
