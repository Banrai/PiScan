// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package digest provides methods for hmac creation and checking, so
// that the client can present its requests with a digest that can be
// authenticated and evaluated by the server

package digest

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
)

func GenerateDigest(privateKey string, message string) string {
	h := hmac.New(sha512.New, []byte(privateKey))
	h.Write([]byte(message))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func DigestMatches(privateKey, messagePlainText, messageDigest string) bool {
	return (messageDigest == GenerateDigest(privateKey, messagePlainText))
}
