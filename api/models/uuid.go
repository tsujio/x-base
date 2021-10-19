package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) MarshalJSON() ([]byte, error) {
	s := uuid.UUID(u)
	str := "\"" + s.String() + "\""
	return []byte(str), nil
}

func (u *UUID) UnmarshalJSON(b []byte) error {
	s, err := uuid.ParseBytes(b)
	if err != nil {
		return err
	}
	*u = UUID(s)
	return nil
}

func (UUID) GormDataType() string {
	return "binary(16)"
}

func (u *UUID) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed to scan value (%v) as uuid", value)
	}
	data, err := uuid.FromBytes(b)
	if err != nil {
		return err
	}
	*u = UUID(data)
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	return uuid.UUID(u).MarshalBinary()
}
