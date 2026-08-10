CREATE TABLE IF NOT EXISTS derivation_briefing_message_operations (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL UNIQUE,
    preparation_session_id TEXT NOT NULL,
    state TEXT NOT NULL,
    failure_code TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);

CREATE TABLE IF NOT EXISTS derivation_briefing_messages (
    preparation_session_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (preparation_session_id, sequence_no),
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id)
);

CREATE TABLE IF NOT EXISTS derivation_briefing_suggestions (
    id TEXT PRIMARY KEY,
    preparation_session_id TEXT NOT NULL,
    operation_id TEXT NOT NULL,
    version_no INTEGER NOT NULL,
    purpose TEXT NOT NULL,
    decision TEXT NOT NULL,
    hypothesis TEXT,
    candidate_prompts TEXT NOT NULL,
    evaluation_criteria TEXT NOT NULL,
    environment_conditions TEXT NOT NULL,
    initial_input TEXT NOT NULL,
    success_criteria TEXT NOT NULL,
    required_conditions TEXT NOT NULL,
    open_question TEXT,
    created_at TEXT NOT NULL,
    UNIQUE (preparation_session_id, version_no),
    FOREIGN KEY (preparation_session_id) REFERENCES preparation_sessions(id),
    FOREIGN KEY (operation_id) REFERENCES derivation_briefing_message_operations(id)
);
