ALTER TABLE experiment_runs ADD COLUMN failure_code TEXT;
ALTER TABLE experiment_runs ADD COLUMN operation_id TEXT;
ALTER TABLE experiment_runs ADD COLUMN snapshot_digest TEXT;
ALTER TABLE experiment_runs ADD COLUMN isolation_kind TEXT;
ALTER TABLE experiment_runs ADD COLUMN last_observed_at TEXT;
ALTER TABLE experiment_runs ADD COLUMN reconciliation_state TEXT;

CREATE TABLE IF NOT EXISTS experiment_run_observations (
    run_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    kind TEXT NOT NULL,
    occurred_at TEXT NOT NULL,
    summary TEXT NOT NULL,
    PRIMARY KEY (run_id, sequence_no),
    FOREIGN KEY (run_id) REFERENCES experiment_runs(id)
);

CREATE TABLE IF NOT EXISTS experiment_run_artifacts (
    run_id TEXT NOT NULL,
    digest TEXT NOT NULL,
    label TEXT,
    status TEXT NOT NULL,
    PRIMARY KEY (run_id, digest),
    FOREIGN KEY (run_id) REFERENCES experiment_runs(id)
);

CREATE TABLE IF NOT EXISTS experiment_run_failures (
    run_id TEXT PRIMARY KEY,
    code TEXT NOT NULL,
    occurred_at TEXT NOT NULL,
    partial_summary TEXT,
    FOREIGN KEY (run_id) REFERENCES experiment_runs(id)
);
