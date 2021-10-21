package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type TableFilesystemEntry struct {
	ID             UUID
	OrganizationID UUID
	Name           string
	Type           string
	ParentFolderID *UUID
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
	ID   UUID
	Type string
	Name string
}

func (e *TableFilesystemEntry) GetPath(db *gorm.DB) ([]TableFilesystemPathEntry, error) {
	tableName := e.Type + "s"
	var entries []TableFilesystemPathEntry
	err := db.Raw(fmt.Sprintf(`
	WITH recursive rec(id, organization_id, type, name, parent_folder_id, depth) AS (
	    SELECT id, organization_id, type, name, parent_folder_id, 0
	    FROM %s
	    WHERE id = ?
	    UNION ALL
	    SELECT f.id, f.organization_id, f.type, f.name, f.parent_folder_id, rec.depth - 1
	    FROM rec
	    INNER JOIN folders AS f
	    ON rec.organization_id = f.organization_id AND
	       rec.parent_folder_id = f.id
	    WHERE rec.parent_folder_id IS NOT NULL
	)

	SELECT id, type, name
	FROM rec
	ORDER BY depth ASC
	`, tableName), e.ID).Scan(&entries).Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get path: %w", err)
	}

	return entries, nil
}
