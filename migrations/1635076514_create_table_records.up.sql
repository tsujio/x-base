CREATE TABLE IF NOT EXISTS table_records (
    id BINARY(16) NOT NULL,
    id_string CHAR(36) AS (BIN_TO_UUID(id)) STORED NOT NULL,
    table_id BINARY(16) NOT NULL,
    data JSON NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_table_records_01 FOREIGN KEY (table_id) REFERENCES tables(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT UNIQUE uq_table_records_01 (id_string),
    INDEX idx_table_records_01 (table_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
