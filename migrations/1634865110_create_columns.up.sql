CREATE TABLE IF NOT EXISTS columns (
    id BINARY(16) NOT NULL,
    table_id BINARY(16) NOT NULL,
    `index` INT UNSIGNED NOT NULL,
    properties JSON NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_columns_01 FOREIGN KEY (table_id) REFERENCES tables(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT uq_columns_01 UNIQUE (table_id, `index`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
