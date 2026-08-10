CREATE TABLE IF NOT EXISTS experiment_derived_operations (
    request_id TEXT PRIMARY KEY,
    source_experiment_id TEXT NOT NULL,
    experiment_id TEXT NOT NULL UNIQUE,
    canonical_payload TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (source_experiment_id) REFERENCES experiments(id),
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);
