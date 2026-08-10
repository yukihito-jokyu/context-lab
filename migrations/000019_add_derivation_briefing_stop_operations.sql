CREATE TABLE IF NOT EXISTS derivation_briefing_stop_operations (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL UNIQUE,
    preparation_session_id TEXT NOT NULL,
    state TEXT NOT NULL,
    failure_code TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);
