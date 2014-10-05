// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// This is a fully-functional (but simple) PiScanner application. So far, all
// it does is define a function which takes the scanned barcode result and
// prints it to stdout. But in time, this binary will grow to do much more...

package main

import (
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/scanner"
	"github.com/Banrai/PiScan/server/commerce/amazon"
	"log"
)

func main() {
	var device string
	flag.StringVar(&device, "device", scanner.SCANNER_DEVICE, fmt.Sprintf("The '/dev/input/event' device associated with your scanner (defaults to '%s')", scanner.SCANNER_DEVICE))
	flag.Parse()

	printFn := func(barcode string) {
		// print the barcode returned by the scanner to stdout
		fmt.Println(fmt.Sprintf("barcode: %s", barcode))

		// and, as a glimpse into the future...
		// lookup the barcode on Amazon's API
		// and print the (json) result to stdout
		// (in the future, this will be handled more elegantly/correctly)
		js, err := amazon.Lookup(barcode)
		if err != nil {
			fmt.Println(fmt.Sprintf("Amazon lookup error: %s", err))
		} else {
			fmt.Println(fmt.Sprintf("Amazon result: %s", js))
		}
	}
	errorFn := func(e error) {
		log.Fatal(e)
	}
	scanner.ScanForever(device, printFn, errorFn)
}
