package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type TableFilesystemEntry struct {
	ID             UUID
	OrganizationID UUID
	Type           string
	ParentFolderID *UUID
	Path           []TableFilesystemPathEntry `gorm:"-"`
	Properties     Properties
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (e *TableFilesystemEntry) BeforeSave(*gorm.DB) error {
	if e.ParentFolderID != nil && *e.ParentFolderID == UUID(uuid.Nil) {
		e.ParentFolderID = nil
	}
	return nil
}

func (e *TableFilesystemEntry) GetTable(db *gorm.DB) (*Table, error) {
	return (&Table{TableFilesystemEntry: *e}).Get(db)
}

func (e *TableFilesystemEntry) GetFolder(db *gorm.DB) (*Folder, error) {
	return (&Folder{TableFilesystemEntry: *e}).Get(db)
}

type TableFilesystemPathEntry struct {
	ID         UUID
	Type       string
	Properties Properties
}

func (e *TableFilesystemEntry) ComputePath(db *gorm.DB) error {
	var entries []TableFilesystemPathEntry
	err := db.Raw(`
	WITH recursive rec(id, organization_id, type, parent_folder_id, properties, depth, all_ids) AS (
	    SELECT id, organization_id, type, parent_folder_id, properties, 0, JSON_ARRAY(id)
	    FROM table_filesystem_entries
	    WHERE id = ?
	    UNION ALL
	    SELECT e.id, e.organization_id, e.type, e.parent_folder_id, e.properties, rec.depth - 1, JSON_ARRAY_APPEND(rec.all_ids, '$', e.id)
	    FROM rec
	    INNER JOIN folders AS f
	    ON f.id = rec.parent_folder_id
	    INNER JOIN table_filesystem_entries AS e
	    ON e.id = f.id AND
	       e.organization_id = rec.organization_id
	    WHERE rec.parent_folder_id IS NOT NULL AND
	          NOT JSON_CONTAINS(rec.all_ids, CAST(e.id AS JSON), '$')
	)
	SELECT id, type, properties
	FROM rec
	WHERE depth != 0
	ORDER BY depth ASC
	`, e.ID).Scan(&entries).Error
	if err != nil {
		return xerrors.Errorf("Failed to get path: %w", err)
	}

	e.Path = entries
	return nil
}
