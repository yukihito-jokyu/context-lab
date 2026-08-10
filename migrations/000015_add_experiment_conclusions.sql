CREATE TABLE IF NOT EXISTS experiment_conclusions (
    id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL UNIQUE,
    conclusion TEXT NOT NULL,
    evaluation_snapshot_digest TEXT NOT NULL,
    state TEXT NOT NULL,
    finalized_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);

CREATE TABLE IF NOT EXISTS experiment_conclusion_operations (
    request_id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL,
    conclusion_id TEXT NOT NULL,
    conclusion TEXT NOT NULL,
    evaluation_snapshot_digest TEXT NOT NULL,
    finalized_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id),
    FOREIGN KEY (conclusion_id) REFERENCES experiment_conclusions(id)
);
