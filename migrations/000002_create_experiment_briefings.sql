CREATE TABLE IF NOT EXISTS preparation_sessions (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS briefing_operations (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL UNIQUE,
    preparation_session_id TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL,
    failure_code TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);
