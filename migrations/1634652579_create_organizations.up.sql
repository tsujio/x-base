CREATE TABLE IF NOT EXISTS organizations (
    id BINARY(16) NOT NULL,
    properties JSON NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
