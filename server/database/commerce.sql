
-- Additional Tables for the POD database:
-- used to track user contributions to the core POD data

-- n.b. this schema uses uuid as binary for the primary keys,
-- since mysql's support for varchar/strings as primaty keys
-- is poor compared to binary fields
-- source: http://stackoverflow.com/a/10951183


-- Commerce: support for scan-to-buy actions from the UI

-- `amazon` defines the product information according to 
-- what's needed to place an order there, i.e., primarily
-- an ASIN code for each barcode

CREATE TABLE amazon (
	id         binary(16) primary key NOT NULL,
	barcode    varchar(13) NOT NULL, -- either GTIN.GTIN_CD (POD) or barcode.barcode (user-contributed)
	asin       varchar(10) NOT NULL, -- the corresponding Amazon product code, as selected by the contributing user
	is_upc     boolean DEFAULT false,
	is_ean     boolean DEFAULT false,
	is_isbn    boolean DEFAULT false,
	posted     datetime, -- automatically filled in by trigger, below
	UNIQUE(barcode, asin)
);

CREATE TRIGGER amazon_on_insert BEFORE INSERT ON `amazon`
    FOR EACH ROW SET NEW.posted = IFNULL(NEW.posted, NOW());

