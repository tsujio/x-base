package models

import (
	"fmt"
	"strings"

	"github.com/tsujio/x-base/api/utils/arrays"
	utilstrings "github.com/tsujio/x-base/api/utils/strings"
)

type GetListSortKey struct {
	Key         string
	OrderAsc    bool
	OrderDesc   bool
	OrderValues []string
}

func convertGetListSortKeyToOrderString(sortKeys []GetListSortKey, sortableKeys []string, categoricalKeys map[string][]string) (string, error) {
	if len(sortKeys) == 0 {
		return "id ASC", nil
	}

	var orders []string
	for _, s := range sortKeys {
		key := utilstrings.ToSnakeCase(s.Key)

		if len(s.OrderValues) > 0 {
			values, exists := categoricalKeys[key]
			if !exists {
				return "", fmt.Errorf("Invalid sort option (not a categorical key)")
			}
			var cases string
			for i, v := range s.OrderValues {
				if !arrays.StringSliceContains(values, v) {
					return "", fmt.Errorf("Invalid sort option (value list)")
				}
				cases += fmt.Sprintf(" WHEN %s = '%s' THEN %d ", key, v, i)
			}
			orders = append(orders, fmt.Sprintf("CASE %s ELSE %d END ASC", cases, len(s.OrderValues)))
			continue
		}

		if arrays.StringSliceContains(sortableKeys, key) {
			var o string
			if s.OrderAsc {
				o = "ASC"
			} else if s.OrderDesc {
				o = "DESC"
			} else {
				return "", fmt.Errorf("Invalid sort option (expected 'asc' or 'desc')")
			}
			orders = append(orders, fmt.Sprintf("%s %s", key, o))
			continue
		}

		return "", fmt.Errorf("Invalid sort key: %s", s.Key)
	}

	return strings.Join(orders, ", "), nil
}
