# PiScan

## About

This is a personal shopping and inventory-tracking device based on the [Raspberry Pi](http://www.raspberrypi.org/) and off-the-shelf [usb](https://en.wikipedia.org/wiki/USB) barcode scanners, with an option to share and contribute to the [Open Product Data](http://product.okfn.org/) (POD) database of product barcodes, as part of the [Saruzai Open Data](https://saruzai.com/) project.

## Installation

### Quickstart (for the impatient)

Copy or download the [pre-built ARM binaries](client#a-install-the-client-binaries) onto your Raspberry Pi (runnning the [Raspbian OS](http://www.raspberrypi.org/downloads/)), and complete the [post-install instructions](client#post-install-configuration).

### More detailed instructions

1. Build your [Raspberry Pi client](client/README.md) from the ground up, either by:
    * [building from source](client#b-install-from-source); or
    * using the [pre-built ARM binaries](client#a-install-the-client-binaries) 

2. [Run your own API server](server/README.md) *(optional, if you do not wish to share and contribute to the barcode data project)*

## Usage

[screenshots + video]

### Acknowledgements

 - Github user [danslimmon](https://github.com/danslimmon) for his [oscar](https://github.com/danslimmon/oscar) project, which inspired this one
 - [Vojtech Pavlik](http://atrey.karlin.mff.cuni.cz/~vojtech) for creating the [Linux Input Driver](http://atrey.karlin.mff.cuni.cz/~vojtech/input/) project
 - [linuxquestions.org](http://www.linuxquestions.org) user <tt>bricedebrignaisplage</tt> for his [post explaining how to read input devices](http://www.linuxquestions.org/questions/programming-9/read-from-a-usb-barcode-scanner-that-simulates-a-keyboard-495358/#post2767643)
 - Github user [gvalkov](https://github.com/gvalkov) for [golang-evdev](https://github.com/gvalkov/golang-evdev) which proved invaluable in implementing the [input_event struct](https://www.kernel.org/doc/Documentation/input/input.txt) in Go
 - Github user [rmulley](https://github.com/rmulley) for [this gist](https://gist.github.com/rmulley/6603544) which was helpful in creating the [emailer package](server/emailer/emailer.go)
 - [Russ Cox](http://research.swtch.com/) for his [clear example](http://play.golang.org/p/V94BPN0uKD) of using Go [template functions](http://golang.org/pkg/text/template/)
 - [Dave Cheney](http://dave.cheney.net/) for his article on [installing Go on the Raspberry Pi](http://dave.cheney.net/2012/09/25/installing-go-on-the-raspberry-pi)
