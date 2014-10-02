
-- Additional Tables for the POD database:
-- used to track user contributions to the core POD data

-- n.b. this schema uses uuid as binary for the primary keys,
-- since mysql's support for varchar/strings as primaty keys
-- is poor compared to binary fields
-- source: http://stackoverflow.com/a/10951183


-- (books are in their own universe)
-- so, later, when find good isbn source, add tables for:
-- BOOK, AUTHOR, BOOK_AUTHORS
