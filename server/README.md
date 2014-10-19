The server components run on a separate machine/network and support the functions of the Raspberry Pi scanner, particularly the barcode lookup.

The server runs a [MySQL](http://www.mysql.com/) database, which is a clone of the current [Open Product Data](http://www.product-open-data.com/download/) (POD) data, with [additional tables](database) for logging user contributions.

## Installation

1. The server code uses this MySQL driver for Go:

   ```sh
go get github.com/go-sql-driver/mysql
   ````