CREATE TABLE IF NOT EXISTS folders (
    id BINARY(16) NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_folders_01 FOREIGN KEY (id) REFERENCES table_filesystem_entries(id) ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
