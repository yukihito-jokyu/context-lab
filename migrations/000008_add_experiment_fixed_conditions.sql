ALTER TABLE experiments ADD COLUMN fixed_condition_id TEXT;

CREATE TABLE IF NOT EXISTS experiment_fixed_conditions (
    id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL UNIQUE,
    purpose TEXT NOT NULL,
    hypothesis TEXT,
    environment_conditions TEXT NOT NULL,
    initial_input TEXT NOT NULL,
    evaluation_axes TEXT NOT NULL,
    artifact_payload TEXT NOT NULL,
    fixed_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);

CREATE TABLE IF NOT EXISTS experiment_fixed_condition_prompts (
    fixed_condition_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    content TEXT NOT NULL,
    PRIMARY KEY (fixed_condition_id, sequence_no),
    FOREIGN KEY (fixed_condition_id) REFERENCES experiment_fixed_conditions(id)
);

CREATE TABLE IF NOT EXISTS experiment_condition_fix_operations (
    request_id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL,
    fixed_condition_id TEXT NOT NULL,
    operation_id TEXT NOT NULL UNIQUE,
    fixed_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id),
    FOREIGN KEY (fixed_condition_id) REFERENCES experiment_fixed_conditions(id)
);
