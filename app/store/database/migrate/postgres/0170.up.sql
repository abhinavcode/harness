-- Add deleted_at and deleted_by columns to registries, images, and artifacts tables
-- Add to registries table
ALTER TABLE registries
    ADD COLUMN IF NOT EXISTS registry_deleted_at BIGINT,
    ADD COLUMN IF NOT EXISTS registry_deleted_by INTEGER;

CREATE INDEX IF NOT EXISTS idx_registries_deleted_at ON registries(registry_deleted_at);

-- Add to images table
ALTER TABLE images
    ADD COLUMN IF NOT EXISTS image_deleted_at BIGINT,
    ADD COLUMN IF NOT EXISTS image_deleted_by INTEGER;

CREATE INDEX IF NOT EXISTS idx_images_deleted_at ON images(image_deleted_at);

-- Add to artifacts table
ALTER TABLE artifacts
    ADD COLUMN IF NOT EXISTS artifact_deleted_at BIGINT,
    ADD COLUMN IF NOT EXISTS artifact_deleted_by INTEGER;

CREATE INDEX IF NOT EXISTS idx_artifacts_deleted_at ON artifacts(artifact_deleted_at);
