ALTER TABLE experiment_evaluations ADD COLUMN run_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_experiment_evaluations_run_id
    ON experiment_evaluations(run_id)
    WHERE run_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS experiment_evaluation_operations (
    request_id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    evaluation_id TEXT NOT NULL UNIQUE,
    operation_id TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL,
    failure_code TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (run_id) REFERENCES experiment_runs(id),
    FOREIGN KEY (evaluation_id) REFERENCES experiment_evaluations(id)
);
