
-- Additional Tables for the POD database:
-- used to track user contributions to the core POD data

-- n.b. this schema uses uuid as binary for the primary keys,
-- since mysql's support for varchar/strings as primaty keys
-- is poor compared to binary fields
-- source: http://stackoverflow.com/a/10951183


-- Contributor management

-- `account` defines the person contributing product information

CREATE TABLE account (
	id            binary(16) primary key NOT NULL,
	email         varchar(512) NOT NULL,
	date_joined   datetime, -- automatically filled in by trigger, below
	verify_code   varchar(32),
	verified      boolean DEFAULT false,
	date_verified datetime,
	enabled       boolean DEFAULT false,
	UNIQUE(email, id)
);

CREATE TRIGGER account_on_insert BEFORE INSERT ON `account`
    FOR EACH ROW SET NEW.date_joined = IFNULL(NEW.date_joined, NOW());

