CREATE TABLE IF NOT EXISTS table_filesystem_entries (
    id BINARY(16) NOT NULL,
    organization_id BINARY(16) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type CHAR(16) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_table_filesystem_entries_01 FOREIGN KEY (organization_id) REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
