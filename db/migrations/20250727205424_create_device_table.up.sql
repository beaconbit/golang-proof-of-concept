CREATE TABLE devices (
    mac TEXT PRIMARY KEY,
    ip TEXT NOT NULL,
    valid BOOLEAN NOT NULL DEFAULT true,
    failures INTEGER NOT NULL DEFAULT 0,
    username TEXT,
    password TEXT,
    cookie TEXT,
    cookie_expires BIGINT NOT NULL DEFAULT 0,
    auth_flow TEXT,
    scraper TEXT,
    last_data TEXT,
    last_seen BIGINT
);

