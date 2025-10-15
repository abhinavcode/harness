
-- For registry table
ALTER TABLE registry ADD COLUMN deleted_at BIGINT;
ALTER TABLE registry ADD COLUMN deleted_by BIGINT;
CREATE INDEX IF NOT EXISTS idx_registry_deleted_at ON registry(deleted_at);

-- For image table
ALTER TABLE image ADD COLUMN deleted_at BIGINT;
ALTER TABLE image ADD COLUMN deleted_by BIGINT;
CREATE INDEX IF NOT EXISTS idx_image_deleted_at ON image(deleted_at);

-- For artifact table
ALTER TABLE artifact ADD COLUMN deleted_at BIGINT;
ALTER TABLE artifact ADD COLUMN deleted_by BIGINT;
CREATE INDEX IF NOT EXISTS idx_artifact_deleted_at ON artifact(deleted_at);
