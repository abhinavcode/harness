CREATE TABLE IF NOT EXISTS node_download_delta
(
    id                      BIGSERIAL PRIMARY KEY,
    node_id                 UUID    NOT NULL UNIQUE,
    download_count_delta    BIGINT  NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_node_download_delta_node_id
        FOREIGN KEY (node_id)
            REFERENCES nodes (node_id)
);

CREATE INDEX IF NOT EXISTS idx_node_download_delta_node_id 
    ON node_download_delta (node_id);
