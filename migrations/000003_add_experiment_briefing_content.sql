ALTER TABLE preparation_sessions ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS briefing_messages (
    preparation_session_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (preparation_session_id, sequence_no),
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);

CREATE TABLE IF NOT EXISTS briefing_versions (
    id TEXT PRIMARY KEY,
    preparation_session_id TEXT NOT NULL,
    version_no INTEGER NOT NULL,
    decision TEXT NOT NULL,
    hypothesis TEXT,
    success_criteria TEXT NOT NULL,
    required_conditions TEXT NOT NULL,
    open_question TEXT,
    created_at TEXT NOT NULL,
    UNIQUE (preparation_session_id, version_no),
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);
