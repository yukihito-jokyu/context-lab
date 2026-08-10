CREATE TABLE IF NOT EXISTS experiment_preparation_draft_operations (
    request_id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL,
    purpose TEXT NOT NULL,
    hypothesis TEXT,
    environment_conditions TEXT NOT NULL,
    initial_input TEXT NOT NULL,
    evaluation_axes TEXT NOT NULL,
    prompts_json TEXT NOT NULL,
    saved_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);
