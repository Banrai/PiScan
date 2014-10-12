// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package barcodes provides access to the database holding product data,
// sourced from both from the Open Product Database (POD) and every supported
// commerce API/site

package barcodes

import (
	"fmt"
	"os"
)

const (
	// Barcode Types
	UPC  = "UPC"
	EAN  = "EAN"
	ISBN = "ISBN"
)

// Universally Unique Identifier (UUID) creation

var (
	// Formatting function: take the 16 random bytes and return them as a string
	// of 8-4-4-4-12 tuples, separated by dashes
	DashedUUID = func(b []byte) string {
		return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	}

	// An alternative formatting function: take the 16 random bytes and return
	// them as a single string, with no dashes
	UndashedUUID = func(b []byte) string {
		return fmt.Sprintf("%x", b)
	}
)

// Generate a universally unique identifier (UUID) using the computer's
// /dev/urandom output as a randomizer, returning a string specified by
// the given formatting function
func GenerateUUID(fn func([]byte) string) string {
	f, e := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	defer f.Close()

	if e != nil {
		return ""
	} else {
		b := make([]byte, 16)
		f.Read(b)
		return fn(b)
	}
}
