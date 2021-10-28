package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
)

type Properties map[string]interface{}

var propertiesKeyPattern = regexp.MustCompile(`^\w+$`)

func ValidateProperties(p map[string]interface{}) string {
	for k, _ := range p {
		if !propertiesKeyPattern.MatchString(k) {
			return fmt.Sprintf("Invalid property key: %s", k)
		}
	}
	return ""
}

func (p *Properties) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Invalid type: %v (%T)", value, value)
	}

	var result map[string]interface{}
	err := json.Unmarshal(bytes, &result)
	*p = Properties(result)
	return err
}

func (p Properties) Value() (driver.Value, error) {
	if p == nil {
		p = make(map[string]interface{})
	}

	if result := ValidateProperties(p); result != "" {
		return nil, fmt.Errorf(result)
	}

	var nullKeys []string
	for k, v := range p {
		if v == nil {
			nullKeys = append(nullKeys, k)
		}
	}
	for _, k := range nullKeys {
		delete(p, k)
	}

	return json.Marshal(&p)
}
