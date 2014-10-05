// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package amazon provides a simple interface to the python
// amazon_api_lookup.py script using os/exec and returns the string result
// and error (if any) as-is

package amazon

import (
	"bytes"
	"os/exec"
	"path"
	"runtime"
)

func Lookup(barcode string) (string, error) {

	// find the path of the calling binary
	_, filename, _, _ := runtime.Caller(1)

	// for now, the path to the API lookup script is relative to the source root
	// so pass the barcode string to it as the first command line argument and
	// capture and return the result
	lookupCmd := []string{"python", path.Join(path.Dir(filename), "/server/commerce/amazon/amazon_api_lookup.py"), barcode}
	cmd := exec.Command(lookupCmd[0], lookupCmd[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	return out.String(), err
}
