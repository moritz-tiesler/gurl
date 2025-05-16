CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY,
    short text NOT NULL,
    original text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_urls_short ON urls(short);