
-- Additional Tables for the POD database:
-- used to track user contributions to the core POD data

-- n.b. this schema uses uuid as binary for the primary keys,
-- since mysql's support for varchar/strings as primaty keys
-- is poor compared to binary fields
-- source: http://stackoverflow.com/a/10951183


-- `book` defines basic title and barcode (isbn) information, and the
-- corresponding source

CREATE TABLE book (
	id        binary(16) primary key NOT NULL,
	title     varchar(512) NOT NULL,
	isbn      varchar(13) NOT NULL,
	is_isbn10 boolean DEFAULT false,
	src       varchar(32) NOT NULL DEFAULT 'OL', -- open library
	posted    datetime, -- automatically filled in by trigger, below
	UNIQUE(title, isbn, src) -- an isbn *should* be unique but sources may differ
);

CREATE TRIGGER book_on_insert BEFORE INSERT ON `book`
    FOR EACH ROW SET NEW.posted = IFNULL(NEW.posted, NOW());


-- `author` defines author names, and the corresponding source information

CREATE TABLE author (
	id        binary(16) primary key NOT NULL,
	full_name varchar(512) NOT NULL,
	src       varchar(32) NOT NULL DEFAULT 'OL', -- open library
	src_id    varchar(64) -- a unique identifier from the source (if present)
);


-- `book_author` defines which author(s) are credited with which books

CREATE TABLE book_author (
	id        binary(16) primary key NOT NULL,
	book_id   binary(16) references book(id),
	author_id binary(16) references author(id),
	UNIQUE(book_id, author_id)
);
