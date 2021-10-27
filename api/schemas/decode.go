package schemas

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"golang.org/x/xerrors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
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

type GetListSortKey struct {
	Key         string
	OrderAsc    bool
	OrderDesc   bool
	OrderValues []string
}

func DecodeGetListSort(sort string, dest *[]GetListSortKey) error {
	if sort == "" {
		return nil
	}

	var sortKeys []GetListSortKey
	for _, token := range strings.Split(sort, ",") {
		kv := strings.SplitN(token, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("Invalid sort key format: %s", token)
		}

		k := GetListSortKey{
			Key: kv[0],
		}

		if strings.ToLower(kv[1]) == "asc" {
			k.OrderAsc = true
		} else if strings.ToLower(kv[1]) == "desc" {
			k.OrderDesc = true
		} else if strings.HasPrefix(kv[1], "(") && strings.HasSuffix(kv[1], ")") {
			var values []string
			for _, v := range strings.Split(kv[1][1:len(kv[1])-1], " ") {
				if v != "" {
					values = append(values, v)
				}
			}
			if len(values) == 0 {
				return fmt.Errorf("Invalid sort key value list: %s", token)
			}
			k.OrderValues = values
		} else {
			return fmt.Errorf("Invalid sort key value format: %s", token)
		}

		sortKeys = append(sortKeys, k)
	}

	*dest = sortKeys

	return nil
}
