
-- Drop indexes first
DROP INDEX IF EXISTS idx_registry_deleted_at;
DROP INDEX IF EXISTS idx_image_deleted_at;
DROP INDEX IF EXISTS idx_artifact_deleted_at;

-- Drop columns from registry table
ALTER TABLE registry DROP COLUMN deleted_at;
ALTER TABLE registry DROP COLUMN deleted_by;

-- Drop columns from image table
ALTER TABLE image DROP COLUMN deleted_at;
ALTER TABLE image DROP COLUMN deleted_by;

-- Drop columns from artifact table
ALTER TABLE artifact DROP COLUMN deleted_at;
ALTER TABLE artifact DROP COLUMN deleted_by;
