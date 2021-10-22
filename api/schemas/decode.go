package schemas

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/tsujio/x-base/api/models"
	"golang.org/x/xerrors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("columntype", func(fl validator.FieldLevel) bool {
		return models.IsValidColumnType(fl.Field().String())
	})
}

func DecodeJSON(source io.Reader, dest interface{}) error {
	err := json.NewDecoder(source).Decode(dest)
	if err != nil {
		return xerrors.Errorf("Failed to decode json: %w", err)
	}
	err = validate.Struct(dest)
	if err != nil {
		return xerrors.Errorf("Invalid input: %w", err)
	}
	return nil
}

func DecodeQuery(source map[string][]string, dest interface{}) error {
	decoder := schema.NewDecoder()
	err := decoder.Decode(dest, source)
	if err != nil {
		return xerrors.Errorf("Failed to decode query: %w", err)
	}
	err = validate.Struct(dest)
	if err != nil {
		return xerrors.Errorf("Invalid input: %w", err)
	}
	return nil
}

func DecodeUUID(source map[string]string, key string, dest *uuid.UUID) error {
	s, exists := source[key]
	if !exists {
		return fmt.Errorf("Key '%s' not exist", key)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return xerrors.Errorf("Failed to parse uuid: %w", err)
	}
	*dest = id
	return nil
}
