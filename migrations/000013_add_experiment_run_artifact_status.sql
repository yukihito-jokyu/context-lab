ALTER TABLE experiment_runs ADD COLUMN artifact_status TEXT NOT NULL DEFAULT 'notRecorded' CHECK (artifact_status IN ('complete', 'partial', 'notRecorded'));
ALTER TABLE experiment_runs ADD COLUMN artifact_reason_code TEXT;
