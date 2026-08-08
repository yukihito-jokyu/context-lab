CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS experiments (
    id TEXT PRIMARY KEY,
    purpose TEXT NOT NULL,
    state TEXT NOT NULL,
    progress_summary TEXT NOT NULL DEFAULT '',
    derived_from_experiment_id TEXT,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS application_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
