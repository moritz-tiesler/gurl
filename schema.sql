-- CREATE TABLE IF NOT EXISTS authors (
--     id INTEGER PRIMARY KEY,
--     name text NOT NULL,
--     bio text,
--     birthday date
-- );
DROP TABLE IF EXISTS authors;

CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY,
    original text NOT NULL,
    short text NOT NULL
);