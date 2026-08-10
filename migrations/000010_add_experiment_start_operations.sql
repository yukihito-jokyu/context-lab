CREATE TABLE IF NOT EXISTS experiment_start_operations (
    request_id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL,
    operation_id TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL,
    failure_code TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (experiment_id) REFERENCES experiments(id)
);

ALTER TABLE experiment_runs ADD COLUMN prompt_sequence_no INTEGER;

CREATE UNIQUE INDEX IF NOT EXISTS idx_experiment_runs_operation_prompt
    ON experiment_runs(experiment_id, prompt_sequence_no)
    WHERE prompt_sequence_no IS NOT NULL;
