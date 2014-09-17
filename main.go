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
	"log"
)

func main() {
	var device string
	flag.StringVar(&device, "device", scanner.SCANNER_DEVICE, fmt.Sprintf("The '/dev/input/event' device associated with your scanner (defaults to '%s')", scanner.SCANNER_DEVICE))
	flag.Parse()

	printFn := func(barcode string) {
		fmt.Println(fmt.Sprintf("barcode: %s", barcode))
	}
	errorFn := func(e error) {
		log.Fatal(e)
	}
	scanner.ScanForever(device, printFn, errorFn)
}
