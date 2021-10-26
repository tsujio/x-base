ALTER TABLE table_filesystem_entries ADD COLUMN parent_folder_id BINARY(16) DEFAULT NULL;
ALTER TABLE table_filesystem_entries ADD CONSTRAINT fk_table_filesystem_entries_02 FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON UPDATE CASCADE ON DELETE CASCADE;
