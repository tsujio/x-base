CREATE TABLE IF NOT EXISTS folders (
    id BINARY(16) NOT NULL,
    organization_id BINARY(16) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type CHAR(16) NOT NULL,
    parent_folder_id BINARY(16),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_folders_01 FOREIGN KEY (organization_id) REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_folders_02 FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
