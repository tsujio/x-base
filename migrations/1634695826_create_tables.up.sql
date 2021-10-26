CREATE TABLE IF NOT EXISTS tables (
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_tables_01 FOREIGN KEY (id) REFERENCES table_filesystem_entries(id) ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
