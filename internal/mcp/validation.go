package mcp

import (
	"fmt"
	"strings"
)

// ValidateArgs checks that the provided arguments conform to the schema.
func ValidateArgs(schema JSONSchema, args map[string]any) error {
	// Check required fields
	for _, field := range schema.Required {
		if _, ok := args[field]; !ok {
			return fmt.Errorf("missing required argument: %s", field)
		}
	}

	// Validate types
	for key, prop := range schema.Properties {
		val, ok := args[key]
		if !ok {
			continue
		}
		if err := validateType(prop.Type, key, val); err != nil {
			return err
		}
		if len(prop.Enum) > 0 {
			strVal := fmt.Sprintf("%v", val)
			valid := false
			for _, e := range prop.Enum {
				if strVal == e {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("argument %q value %q is not valid, must be one of: %s", key, strVal, strings.Join(prop.Enum, ", "))
			}
		}
	}

	return nil
}

func validateType(expectedType, key string, val any) error {
	switch expectedType {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("argument %q must be a string, got %T", key, val)
		}
	case "integer":
		switch val.(type) {
		case float64, int, int64:
			// JSON numbers decode as float64
		default:
			return fmt.Errorf("argument %q must be an integer, got %T", key, val)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("argument %q must be a boolean, got %T", key, val)
		}
	case "array":
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("argument %q must be an array, got %T", key, val)
		}
	}
	return nil
}
