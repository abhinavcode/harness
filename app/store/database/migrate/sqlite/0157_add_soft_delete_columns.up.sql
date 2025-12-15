
-- For registries table
ALTER TABLE registries ADD COLUMN registry_deleted_at INTEGER;
ALTER TABLE registries ADD COLUMN registry_deleted_by INTEGER;
CREATE INDEX IF NOT EXISTS idx_registries_deleted_at ON registries(registry_deleted_at);

-- For images table
ALTER TABLE images ADD COLUMN image_deleted_at INTEGER;
ALTER TABLE images ADD COLUMN image_deleted_by INTEGER;
CREATE INDEX IF NOT EXISTS idx_images_deleted_at ON images(image_deleted_at);

-- For artifacts table
ALTER TABLE artifacts ADD COLUMN artifact_deleted_at INTEGER;
ALTER TABLE artifacts ADD COLUMN artifact_deleted_by INTEGER;
CREATE INDEX IF NOT EXISTS idx_artifacts_deleted_at ON artifacts(artifact_deleted_at);
