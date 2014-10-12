// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package commerce provides general objects and functions for any product
// vendor API or website

package commerce

// API represents the minimum required output from any vendor's API
type API struct {
	SKU         string `json:"sku"`
	ProductName string `json:"desc,omitempty"`
	ProductType string `json:"type,omitempty"`
	Vendor      string `json:"vnd"`
}
