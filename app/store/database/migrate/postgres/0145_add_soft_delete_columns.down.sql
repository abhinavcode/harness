-- Remove from artifact table
DROP INDEX IF EXISTS idx_artifact_deleted_at;
ALTER TABLE artifact 
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by;

-- Remove from image table
DROP INDEX IF EXISTS idx_image_deleted_at;
ALTER TABLE image 
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by;

-- Remove from registry table
DROP INDEX IF EXISTS idx_registry_deleted_at;
ALTER TABLE registry 
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by;
