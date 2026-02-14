package schema

import (
	"encoding/json"
	"envdoc/internal/types"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Generate creates a schema from the parsed environment variables
func Generate(envVarMap types.EnvVarMap) types.Schema {
	schema := types.Schema{}

	for key, item := range envVarMap {
		valueType := guessValueType(item.Value)
		schema[key] = types.SchemaItem{
			Example:  item.Value,
			Type:     valueType,
			Required: false,
		}
	}

	return schema
}

// Output formats the schema as JSON, YAML, or text
func Output(schema types.Schema, format string) (string, error) {
	switch format {
	case "json":
		output, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return "", err
		}
		return string(output), nil

	case "yaml":
		output, err := yaml.Marshal(schema)
		if err != nil {
			return "", err
		}
		return string(output), nil

	case "text":
		var sb strings.Builder
		for key, item := range schema {
			sb.WriteString(fmt.Sprintf("%s:\n", key))
			sb.WriteString(fmt.Sprintf("  example: %s\n", item.Example))
			sb.WriteString(fmt.Sprintf("  type: %s\n", item.Type))
			sb.WriteString(fmt.Sprintf("  required: %v\n", item.Required))
			sb.WriteString("\n")
		}
		return sb.String(), nil

	default:
		return "", fmt.Errorf("unknown output format: %s (use json, yaml, or text)", format)
	}
}

func guessValueType(value string) string {
	value = strings.TrimSpace(value)

	if value == "true" || value == "false" {
		return "boolean"
	}
	if _, err := strconv.Atoi(value); err == nil {
		return "integer"
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return "float"
	}
	return "string"
}
