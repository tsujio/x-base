package models

import (
	"fmt"
	"regexp"
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

var categoricalValuePattern = regexp.MustCompile(`^\w+$`)

func convertGetListSortKeyToOrderString(sortKeys []GetListSortKey, sortableKeys []string) (string, error) {
	if len(sortKeys) == 0 {
		return "id ASC", nil
	}

	var orders []string
	for _, s := range sortKeys {
		var key string
		if strings.HasPrefix(s.Key, "property.") {
			prop := s.Key[len("property."):]
			if !propertiesKeyPattern.MatchString(prop) {
				return "", fmt.Errorf("Invalid sort key: %s", s.Key)
			}
			key = fmt.Sprintf("JSON_EXTRACT(properties, '$.%s')", prop)
		} else {
			key = utilstrings.ToSnakeCase(s.Key)
			if !arrays.StringSliceContains(sortableKeys, key) {
				return "", fmt.Errorf("Invalid sort key: %s", s.Key)
			}
		}

		if len(s.OrderValues) > 0 {
			var cases string
			for i, v := range s.OrderValues {
				if !categoricalValuePattern.MatchString(v) {
					return "", fmt.Errorf("Invalid sort option (value list)")
				}
				cases += fmt.Sprintf(" WHEN %s = '%s' THEN %d ", key, v, i)
			}
			orders = append(orders, fmt.Sprintf("CASE %s ELSE %d END ASC", cases, len(s.OrderValues)))
			continue
		}

		var o string
		if s.OrderAsc {
			o = "ASC"
		} else if s.OrderDesc {
			o = "DESC"
		} else {
			return "", fmt.Errorf("Invalid sort option (expected 'asc' or 'desc')")
		}
		orders = append(orders, fmt.Sprintf("%s %s", key, o))
	}

	return strings.Join(orders, ", "), nil
}
