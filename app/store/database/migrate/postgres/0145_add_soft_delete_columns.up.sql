-- Add deleted_at and deleted_by columns to registry, image, and artifact tables

-- Add to registry table
ALTER TABLE registry 
    ADD COLUMN IF NOT EXISTS deleted_at BIGINT,
    ADD COLUMN IF NOT EXISTS deleted_by BIGINT;

CREATE INDEX IF NOT EXISTS idx_registry_deleted_at ON registry(deleted_at);

-- Add to image table
ALTER TABLE image 
    ADD COLUMN IF NOT EXISTS deleted_at BIGINT,
    ADD COLUMN IF NOT EXISTS deleted_by BIGINT;

CREATE INDEX IF NOT EXISTS idx_image_deleted_at ON image(deleted_at);

-- Add to artifact table
ALTER TABLE artifact 
    ADD COLUMN IF NOT EXISTS deleted_at BIGINT,
    ADD COLUMN IF NOT EXISTS deleted_by BIGINT;

CREATE INDEX IF NOT EXISTS idx_artifact_deleted_at ON artifact(deleted_at);
