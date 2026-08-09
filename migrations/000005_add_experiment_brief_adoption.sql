ALTER TABLE briefing_versions ADD COLUMN purpose TEXT NOT NULL DEFAULT '';
ALTER TABLE briefing_versions ADD COLUMN candidate_prompts TEXT NOT NULL DEFAULT '[]';
ALTER TABLE briefing_versions ADD COLUMN evaluation_criteria TEXT NOT NULL DEFAULT '';
ALTER TABLE briefing_versions ADD COLUMN environment_conditions TEXT NOT NULL DEFAULT '';
ALTER TABLE briefing_versions ADD COLUMN initial_input TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS experiment_preparations (
    experiment_id TEXT PRIMARY KEY,
    briefing_session_id TEXT NOT NULL,
    briefing_version_id TEXT NOT NULL,
    hypothesis TEXT,
    environment_conditions TEXT NOT NULL,
    initial_input TEXT NOT NULL,
    evaluation_criteria TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id),
    FOREIGN KEY (briefing_session_id) REFERENCES preparation_sessions(id),
    FOREIGN KEY (briefing_version_id) REFERENCES briefing_versions(id)
);

CREATE TABLE IF NOT EXISTS experiment_preparation_prompts (
    experiment_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    content TEXT NOT NULL,
    PRIMARY KEY (experiment_id, sequence_no),
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);

CREATE TABLE IF NOT EXISTS experiment_creation_operations (
    request_id TEXT PRIMARY KEY,
    briefing_session_id TEXT NOT NULL,
    briefing_version_id TEXT NOT NULL,
    experiment_id TEXT NOT NULL UNIQUE,
    FOREIGN KEY (briefing_session_id) REFERENCES preparation_sessions(id),
    FOREIGN KEY (briefing_version_id) REFERENCES briefing_versions(id),
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS experiment_creation_operations_brief_version
    ON experiment_creation_operations (briefing_session_id, briefing_version_id);
