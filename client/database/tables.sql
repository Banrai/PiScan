
-- These tables comprise the local datastore on the Raspberry Pi client
-- device, using SQLite for the database. SQLite has a limited set of
-- datatypes (https://www.sqlite.org/datatype3.html), so the analogous
-- server database columns have been adjusted accordingly.

-- `account` defines basic end-user information, corresponding to the
-- account table in the server database

CREATE TABLE IF NOT EXISTS account (
	id       integer primary key AUTOINCREMENT,
	email    text NOT NULL,
	api_code text NOT NULL,
	UNIQUE(email)
);

-- `product` defines the items scanned, edited (when the barcode lookup
-- resulted in no matches), and favorited by a given end-user

CREATE TABLE IF NOT EXISTS product (
	id           integer primary key AUTOINCREMENT,
	barcode      text NOT NULL,
	product_desc text, -- can be null: means the scanned item is unknown
	product_ind  integer DEFAULT 0, -- to distinguish multiple products with the same barcode
	is_favorite  integer DEFAULT 0, -- 0 = false, 1 = true
	is_edit      integer DEFAULT 0, -- 0 = false, 1 = true
	posted       integer, -- unix time (i.e., the number of seconds since 1970-01-01 00:00:00 UTC)
	account      integer REFERENCES account
); 

-- `product_availability` defines where a given product can be purchased,
-- according to the commerce-related tables in the server database (the
-- unique list of vendor ids provides the "Buy from ..." action options)

CREATE TABLE IF NOT EXISTS product_availability (
	id        integer primary key AUTOINCREMENT,
	vendor_id text NOT NULL,
	product   integer REFERENCES product
);

