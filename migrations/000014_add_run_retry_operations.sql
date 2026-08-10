DROP INDEX IF EXISTS idx_experiment_runs_operation_prompt;
CREATE INDEX IF NOT EXISTS idx_experiment_runs_operation_prompt
    ON experiment_runs(experiment_id, prompt_sequence_no, created_at, id);

ALTER TABLE experiment_runs ADD COLUMN retry_of_run_id TEXT;

CREATE TABLE IF NOT EXISTS experiment_run_retry_operations (
    request_id TEXT PRIMARY KEY,
    source_run_id TEXT NOT NULL,
    experiment_id TEXT NOT NULL,
    run_id TEXT NOT NULL UNIQUE,
    operation_id TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (source_run_id) REFERENCES experiment_runs(id),
    FOREIGN KEY (run_id) REFERENCES experiment_runs(id)
);
