The client components all run on the Raspberry Pi device, along with the barcode scanner.

The client datastore is a simple [SQLite](http://sqlite.org/) database file, consisting of the [basic tables](database) needed to keep track of individual user scans, product data contributions, and favorited items.

## Installation

1. The client code uses this SQLite driver for Go:

   ```sh
go get github.com/mxk/go-sqlite/sqlite3
   ```