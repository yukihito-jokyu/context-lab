CREATE TABLE IF NOT EXISTS derivation_briefing_operations (
    request_id TEXT PRIMARY KEY,
    source_experiment_id TEXT NOT NULL,
    preparation_session_id TEXT NOT NULL UNIQUE,
    operation_id TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL,
    failure_code TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (source_experiment_id) REFERENCES experiments(id),
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);
