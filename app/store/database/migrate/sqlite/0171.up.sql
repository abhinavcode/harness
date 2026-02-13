CREATE TABLE IF NOT EXISTS node_download_delta
(
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id                 TEXT    NOT NULL UNIQUE,
    download_count_delta    INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_node_download_delta_node_id
        FOREIGN KEY (node_id)
            REFERENCES nodes (node_id)
);

CREATE INDEX IF NOT EXISTS idx_node_download_delta_node_id 
    ON node_download_delta (node_id);
