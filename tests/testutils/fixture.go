package testutils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/tsujio/x-base/api/models"
)

var uuids = map[string]uuid.UUID{}

func GetUUID(name string) uuid.UUID {
	id, exists := uuids[name]
	if !exists {
		id = uuid.New()
		uuids[name] = id
	}
	return id
}

func Dedent(s string) string {
	return regexp.MustCompile(`(?m)^\t+`).ReplaceAllString(s, "")
}

func LoadFixture(yml string) error {
	var fixture map[string]interface{}
	err := yaml.Unmarshal([]byte(Dedent(yml)), &fixture)
	if err != nil {
		return xerrors.Errorf("Failed to parse yaml: %w", err)
	}

	if organizations, exists := fixture["organizations"]; exists {
		if err := createOrganizations(organizations, ".organizations"); err != nil {
			return err
		}
	}

	return nil
}

func createOrganizations(organizations interface{}, path string) error {
	if orgs, ok := organizations.([]interface{}); !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, organizations)
	} else {
		for i, o := range orgs {
			if err := createOrganization(o, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func createOrganization(organization interface{}, path string) error {
	if org, ok := organization.(map[string]interface{}); !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, organization)
	} else {
		o := models.Organization{}

		// ID
		if id, exists := org["id"]; exists {
			if idStr, ok := id.(string); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".id", id)
			} else {
				o.ID = models.UUID(GetUUID(idStr))
			}
		} else {
			o.ID = models.UUID(uuid.New())
		}

		// Properties
		if properties, exists := org["properties"]; exists {
			if props, ok := properties.(map[string]interface{}); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".properties", properties)
			} else {
				o.Properties = props
			}
		}

		// CreatedAt
		if createdAt, exists := org["createdAt"]; exists {
			if createdAtStr, ok := createdAt.(string); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".createdAt", createdAt)
			} else {
				if createdAtTime, err := time.Parse(time.RFC3339, createdAtStr); err != nil {
					return fmt.Errorf("Invalid time format: path=%s", path+".createdAt")
				} else {
					o.CreatedAt = createdAtTime
				}
			}
		}

		if err := o.Create(GetDB()); err != nil {
			return err
		}

		// tables
		if tables, exists := org["tables"]; exists {
			if err := createTableFilesystem(tables, path+".tables", uuid.UUID(o.ID), uuid.Nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func createTableFilesystem(entries interface{}, path string, organizationID, parentFolderID uuid.UUID) error {
	ents, ok := entries.([]interface{})
	if !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, entries)
	}
	for i, e := range ents {
		ent, ok := e.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Invalid type: path=%s[%d], type=%T", path, i, e)
		}

		var typ string
		if tp, exists := ent["type"]; exists {
			if t, ok := tp.(string); !ok {
				return fmt.Errorf("Invalid type: path=%s[%d].type, type=%T", path, i, tp)
			} else {
				typ = t
			}
		} else {
			for _, key := range []string{"id", "name"} {
				if val, exists := ent[key]; exists {
					if v, ok := val.(string); ok {
						for _, t := range []string{"table", "folder"} {
							if strings.HasPrefix(v, t) {
								typ = t
							}
						}
					}
				}
			}
		}

		if typ == "" {
			return fmt.Errorf(".type required: path=%s[%d]", path, i)
		}

		switch typ {
		case "table":
			if err := createTable(e, fmt.Sprintf("%s[%d]", path, i), organizationID, parentFolderID); err != nil {
				return err
			}
		case "folder":
			if err := createFolder(e, fmt.Sprintf("%s[%d]", path, i), organizationID, parentFolderID); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Invalid .type value: path=%s[%d]", path, i)
		}
	}
	return nil
}

func makeTableFilesystemEntry(entry map[string]interface{}, path string, organizationID, parentFolderID uuid.UUID) (*models.TableFilesystemEntry, error) {
	e := &models.TableFilesystemEntry{}

	// ID
	if id, exists := entry["id"]; exists {
		if idStr, ok := id.(string); !ok {
			return nil, fmt.Errorf("Invalid type: path=%s, type=%T", path+".id", id)
		} else {
			e.ID = models.UUID(GetUUID(idStr))
		}
	} else {
		e.ID = models.UUID(uuid.New())
	}

	// OrganizationID
	e.OrganizationID = models.UUID(organizationID)

	// ParentFolderID
	if parentFolderID == uuid.Nil {
		e.ParentFolderID = nil
	} else {
		e.ParentFolderID = (*models.UUID)(&parentFolderID)
	}

	// Properties
	if properties, exists := entry["properties"]; exists {
		if props, ok := properties.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("Invalid type: path=%s, type=%T", path+".properties", properties)
		} else {
			e.Properties = props
		}
	}

	// CreatedAt
	if createdAt, exists := entry["createdAt"]; exists {
		if createdAtStr, ok := createdAt.(string); !ok {
			return nil, fmt.Errorf("Invalid type: path=%s, type=%T", path+".createdAt", createdAt)
		} else {
			if createdAtTime, err := time.Parse(time.RFC3339, createdAtStr); err != nil {
				return nil, fmt.Errorf("Invalid time format: path=%s", path+".createdAt")
			} else {
				e.CreatedAt = createdAtTime
			}
		}
	}

	return e, nil
}

func createTable(table interface{}, path string, organizationID, parentFolderID uuid.UUID) error {
	if tbl, ok := table.(map[string]interface{}); !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, table)
	} else {
		if entry, err := makeTableFilesystemEntry(tbl, path, organizationID, parentFolderID); err != nil {
			return err
		} else {
			t := models.Table{TableFilesystemEntry: *entry}
			if err := t.Create(GetDB()); err != nil {
				return err
			}

			if cols, exists := tbl["columns"]; exists {
				if cs, ok := cols.([]interface{}); !ok {
					return fmt.Errorf("Invalid type: path=%s, type=%T", path+".columns", cols)
				} else {
					for i, c := range cs {
						if err := createColumn(c, fmt.Sprintf("%s.columns[%d]", path, i), t, i); err != nil {
							return err
						}
					}
				}
			}

			if records, exists := tbl["records"]; exists {
				if rcds, ok := records.([]interface{}); !ok {
					return fmt.Errorf("Invalid type: path=%s, type=%T", path+".records", records)
				} else {
					for i, r := range rcds {
						if err := createTableRecord(r, fmt.Sprintf("%s.records[%d]", path, i), t); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func createFolder(folder interface{}, path string, organizationID, parentFolderID uuid.UUID) error {
	if fld, ok := folder.(map[string]interface{}); !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, folder)
	} else {
		if entry, err := makeTableFilesystemEntry(fld, path, organizationID, parentFolderID); err != nil {
			return err
		} else {
			f := models.Folder{TableFilesystemEntry: *entry}
			if err := f.Create(GetDB()); err != nil {
				return err
			}

			if children, exists := fld["children"]; exists {
				if err := createTableFilesystem(children, path+".children", organizationID, uuid.UUID(f.ID)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createColumn(column interface{}, path string, table models.Table, index int) error {
	if col, ok := column.(map[string]interface{}); !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, column)
	} else {
		c := &models.Column{}

		// ID
		if id, exists := col["id"]; exists {
			if idStr, ok := id.(string); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".id", id)
			} else {
				c.ID = models.UUID(GetUUID(idStr))
			}
		} else {
			c.ID = models.UUID(uuid.New())
		}

		// TableID
		c.TableID = table.ID

		// Index
		c.Index = index

		// Properties
		if properties, exists := col["properties"]; exists {
			if props, ok := properties.(map[string]interface{}); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".properties", properties)
			} else {
				c.Properties = props
			}
		}

		if err := c.Create(GetDB(), false); err != nil {
			return err
		}
	}
	return nil
}

func createTableRecord(record interface{}, path string, table models.Table) error {
	if rcd, ok := record.(map[string]interface{}); !ok {
		return fmt.Errorf("Invalid type: path=%s, type=%T", path, record)
	} else {
		var r models.TableRecord

		// ID
		if id, exists := rcd["id"]; exists {
			if idStr, ok := id.(string); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".id", id)
			} else {
				r.ID = models.UUID(GetUUID(idStr))
			}
		} else {
			r.ID = models.UUID(uuid.New())
		}

		// TableID
		r.TableID = table.ID

		// Data
		if err := table.FetchColumns(GetDB()); err != nil {
			return err
		}
		if data, exists := rcd["data"]; !exists {
			return fmt.Errorf(".data required: path=%s", path)
		} else {
			if dt, ok := data.([]interface{}); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".data", data)
			} else {
				m := map[string]interface{}{}
				for i, c := range table.Columns {
					m[c.ID.String()] = dt[i]
				}
				j, err := json.Marshal(&m)
				if err != nil {
					return err
				}
				r.Data = j
			}
		}

		// Properties
		if properties, exists := rcd["properties"]; exists {
			if props, ok := properties.(map[string]interface{}); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".properties", properties)
			} else {
				r.Properties = props
			}
		}

		// CreatedAt
		if createdAt, exists := rcd["createdAt"]; exists {
			if createdAtStr, ok := createdAt.(string); !ok {
				return fmt.Errorf("Invalid type: path=%s, type=%T", path+".createdAt", createdAt)
			} else {
				if createdAtTime, err := time.Parse(time.RFC3339, createdAtStr); err != nil {
					return fmt.Errorf("Invalid time format: path=%s", path+".createdAt")
				} else {
					r.CreatedAt = createdAtTime
				}
			}
		}

		if err := GetDB().Create(&r).Error; err != nil {
			return err
		}
	}
	return nil
}
