ALTER TABLE experiment_evaluations ADD COLUMN snapshot_digest TEXT;
ALTER TABLE experiment_evaluations ADD COLUMN last_observed_at TEXT;
ALTER TABLE experiment_evaluations ADD COLUMN reconciliation_state TEXT;
ALTER TABLE experiment_evaluations ADD COLUMN result_status TEXT;
ALTER TABLE experiment_evaluations ADD COLUMN result_reason_code TEXT;

CREATE TABLE IF NOT EXISTS experiment_evaluation_results (
    evaluation_id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    summary TEXT,
    reason_code TEXT,
    FOREIGN KEY (evaluation_id) REFERENCES experiment_evaluations(id)
);
