ALTER TABLE environment_preparation_operations ADD COLUMN scope TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS environment_preparation_operations_active_scope ON environment_preparation_operations(scope) WHERE state IN ('starting', 'running');
