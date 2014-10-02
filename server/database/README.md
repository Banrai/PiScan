# Barcode Database

The database is clone of the [Open Product Database (POD)](http://www.product-open-data.com/), with additional tables for user-reported contributions, for cases when a given scan is not found in POD, but someone manually identifies it.

The goal is to post these individual contributions back into POD, following the [Open Street Map](http://www.openstreetmap.org/) model.

This barcode database will be hosted and available at [TBD], but this guide is here for anyone who wants to run it themselves.

*n.b.* the barcode database should *not* be run on the Raspberry Pi device controlling the scanner.

## Installation

1. Install [mysql](http://www.mysql.com/)

   On debian/ubuntu, use the built-in package manager:

   ```sh 
sudo apt-get install -y mysql-server
```

2. Download the POD data

   The latest version is [pod_web_2014.01.01_01.sql.gz](http://www.product-open-data.com/docs/pod_web_2014.01.01_01.sql.gz)

   ```sh 
wget http://www.product-open-data.com/docs/pod_web_2014.01.01_01.sql.gz
gunzip pod_web_2014.01.01_01.sql.gz
```

3. Create the database and import the download

   ```sh
mysqladmin -u root create product_open_data
mysql -u root product_open_data < pod_web_2014.01.01_01.sql
```

4. Install the additional tables from this folder

   ```sh
cd $GOPATH/src/github.com/Banrai/PiScan/server/database
mysql -u root product_open_data < contributors.sql
mysql -u root product_open_data < products.sql
mysql -u root product_open_data < books.sql
mysql -u root product_open_data < commerce.sql
```
