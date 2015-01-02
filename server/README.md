The server components run on a separate machine/network and support the functions of the Raspberry Pi scanner, particularly the barcode lookup.

The server runs a [MySQL](http://www.mysql.com/) database, which is a clone of the last available [Open Product Data](http://product.okfn.org/) (POD) data, with [additional tables](database) for logging user contributions.

## Initial server preparation

1. Basic packages using [apt](http://linux.die.net/man/8/apt-get) as root/sudo:

  ```sh
  # apt-get update; apt-get upgrade -y
  # apt-get install -y git mercurial build-essential python-all-dev python-lxml python-magic python-pip python-pycurl screen
  # pip install python-amazon-product-api
  # adduser pod 
  ```

2. POD-specific setup as the <tt>pod</tt> user:

  ```sh
  $ mkdir data
  $ cd data
  $ wget http://www.product-open-data.com/docs/pod_web_2014.01.01_01.sql.gz
  $ gunzip pod_web_2014.01.01_01.sql.gz
  $ mysqladmin -u root -p create product_open_data
  $ mysql -u root -p product_open_data < pod_web_2014.01.01_01.sql
  $ mysql -u root -p product_open_data < piscan_tables.sql
  $ mysql -u root -p product_open_data < piscan_user.sql
  $ mysql -u pod product_open_data
  ```

  The <tt>piscan_tables.sql</tt> file is a convenience (in lieu of invoking <tt>mysql</tt> over each sql file) prepared by combining all the [additional tables](database) from this repo, like this:

  ```sh
$ cat contributors.sql products.sql commerce.sql books.sql >piscan_tables.sql
  ```

  The <tt>piscan_user.sql</tt> file is defined as:

  ```sql
CREATE USER 'pod'@'localhost';
GRANT ALL PRIVILEGES ON *.* TO 'pod'@'localhost'
WITH GRANT OPTION;
FLUSH PRIVILEGES;
  ```

## Server application installation

[to be continued]