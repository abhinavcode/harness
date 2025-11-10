-- Remove from artifacts table
DROP INDEX IF EXISTS idx_artifact_deleted_at;
ALTER TABLE artifacts 
    DROP COLUMN IF EXISTS artifact_deleted_at,
    DROP COLUMN IF EXISTS artifact_deleted_by;

-- Remove from images table
DROP INDEX IF EXISTS idx_image_deleted_at;
ALTER TABLE images 
    DROP COLUMN IF EXISTS image_deleted_at,
    DROP COLUMN IF EXISTS image_deleted_by;

-- Remove from registries table
DROP INDEX IF EXISTS idx_registry_deleted_at;
ALTER TABLE registries 
    DROP COLUMN IF EXISTS registry_deleted_at,
    DROP COLUMN IF EXISTS registry_deleted_by;
