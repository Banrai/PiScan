The server components run on a separate machine/network and support the functions of the Raspberry Pi scanner, particularly the barcode lookup.

By default, the Pi client talks to the API service on the [saruzai.com](http://saruzai.com/) server, which runs a [MySQL](http://www.mysql.com/) database clone of the last available [Open Product Data](http://product.okfn.org/) (POD) data, along with [additional tables](database) for logging user contributions.

If you do not wish to participate in that project, you can build and run your own server instead, by following these instructions.

## Initial server preparation

These instructions assume the underlying OS is [debian linux](http://debian.org/). If you want to use something else, you will need to make the necessary pkg and configuration changes.

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

  *Note that while full POD database file (<tt>pod_web_2014.01.01_01.sql.gz</tt>) was available until recently at <tt>http://www.product-open-data.com/</tt> that site has gone dark suddenly, and it now points to a suspicious site with a bad certificate. If you want a copy of the original POD database without any user contributions from the Saruzai Open Data project, please [contact us](http://banrai.com/contact.html).*

## Server application installation

*These steps are for the [saruzai.com](http://saruzai.com/) server; please change your domain or IP address accordingly.*

1. Prepare the server environment

  As the <tt>pod</tt> user:

  ```sh
$ cd /home/pod
$ mkdir -p server/commerce/amazon/
  ```

2. Copy the Amazon API scripts

  ```sh
scp server/commerce/amazon/*.py pod@saruzai.com:/home/pod/server/commerce/amazon
  ```

  Optionally, login as the <tt>pod</tt> user and create a <tt>amazon_local_settings.py</tt> file, which defines your own Amazon API coordinates.

  ```sh
$ cd server/commerce/amazon/
$ vi amazon_local_settings.py
  ```

  As [described here](commerce/amazon/amazon_settings.py), the <tt>amazon_local_settings.py</tt> file overrides the default (empty) settings for the access and secret keys.

3. Download the [APIServer](https://www.dropbox.com/s/gnwgmdsnrtmrmlu/APIServer?dl=0) binary to the server 

  ```sh
scp APIServer pod@saruzai.com:/home/pod/server
  ```

  *This link is a <tt>linux/amd64</tt> binary. If your server is different, you must build from source instead.*

4. APIServer startup script

  Copy the [api-server.sh](init.d/api-server.sh) script to the server:
 
  ```sh
cd server/init.d
scp api-server.sh pod@saruzai.com:/tmp                                                     
  ```

  Then, as root/sudo, move it to <tt>/etc/init.d</tt> with the correct permissions, and install via <tt>update-rc.d</tt> to start automatically on boot:

  ```sh
# mv /tmp/api-server.sh /etc/init.d
# chmod 755 /etc/init.d/api-server.sh
# update-rc.d api-server.sh defaults
  ```
