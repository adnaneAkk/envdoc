package main

import (
	"bufio"
	"encoding/json"
	"envdoc/types"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Global flags
var (
	strict       bool
	outputFormat string
	outputFile   string
	envFile      string
)

// Root command - default validation
var rootCmd = &cobra.Command{
	Use:   "envdoc [.env file]",
	Short: "Parse and validate .env files",
	Long:  `A fast and flexible .env file parser with schema generation and validation`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// see if you wanna use default or custom .env
		if len(args) > 0 {
			envFile = args[0]
		} else {
			envFile = ".env"
		}

		runValidation(envFile, strict)
	},
}

// Schema generation command
var schemaCmd = &cobra.Command{
	Use:   "schema [.env file]",
	Short: "Generate schema from .env file",
	Long:  `Generate a JSON or YAML schema documenting all environment variables`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine which .env file to use
		if len(args) > 0 {
			envFile = args[0]
		} else {
			envFile = ".env"
		}

		runSchemaGeneration(envFile, strict, outputFormat, outputFile)
	},
}

func init() {
	// Global flags (available to all commands)
	rootCmd.PersistentFlags().BoolVarP(&strict, "strict", "s", false, "Enable strict mode")

	// specific flags for the schema command
	schemaCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json|yaml|text)")
	schemaCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	rootCmd.AddCommand(schemaCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// validates .env file and reports errors/warnings
func runValidation(filename string, strictMode bool) {
	config := types.Config{
		Strict:         strictMode,
		GenerateSchema: false,
	}

	envVarMap, errors, warnings := parseEnvFile(filename, config)

	fmt.Printf("Errors: %d found\n", len(errors))
	for _, e := range errors {
		fmt.Printf("Line %d [%s]: %s (Key: %s)\n", e.LineNum, e.IssueType, e.Message, e.KeyName)
	}

	fmt.Printf("\nWarnings: %d found\n", len(warnings))
	for _, w := range warnings {
		fmt.Printf("Line %d [%s]: %s (Key: %s)\n", w.LineNum, w.IssueType, w.Message, w.KeyName)
	}

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Printf("\n✓ %s is valid! Found %d environment variables.\n", filename, len(envVarMap))
	}

	if len(errors) > 0 {
		os.Exit(1)
	}
}

// generates schema from .env file
func runSchemaGeneration(filename string, strictMode bool, format string, outFile string) {
	config := types.Config{
		Strict:         strictMode,
		GenerateSchema: true,
	}

	envVarMap, errors, warnings := parseEnvFile(filename, config)

	// showing errors THEN warnings
	if len(errors) > 0 {
		fmt.Printf("Errors: %d found\n", len(errors))
		for _, e := range errors {
			fmt.Printf("Line %d [%s]: %s (Key: %s)\n", e.LineNum, e.IssueType, e.Message, e.KeyName)
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %d found\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("Line %d [%s]: %s (Key: %s)\n", w.LineNum, w.IssueType, w.Message, w.KeyName)
		}
	}

	// Building my schema
	schema := types.Schema{}
	for key, item := range envVarMap {
		valueType := GuessValueType(item.Value)
		schema[key] = types.SchemaItem{
			Example:  item.Value,
			Type:     valueType,
			Required: false,
		}
	}

	// Generate output based on format
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(schema, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(schema)
	case "text":
		// Text format kinda just for development for now
		for key, item := range schema {
			fmt.Printf("%s:\n", key)
			fmt.Printf("  example: %s\n", item.Example)
			fmt.Printf("  type: %s\n", item.Type)
			fmt.Printf("  required: %v\n", item.Required)
		}
		return
	default:
		log.Fatalf("Unknown output format: %s (use json, yaml, or text)", format)
	}

	if err != nil {
		log.Fatalf("Error generating output: %v", err)
	}

	// Write to file or stdout
	if outFile != "" {
		if err := os.WriteFile(outFile, output, 0644); err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
		fmt.Printf("Schema written to %s\n", outFile)
	} else {
		fmt.Println(string(output))
	}

	if len(errors) > 0 {
		os.Exit(1)
	}
}

// core parsing logic after i exctracted it from old main for the sake of more readability
func parseEnvFile(filename string, config types.Config) (types.EnvVarMap, []types.Issue, []types.Issue) {
	var errors []types.Issue
	var warnings []types.Issue

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file %s: %v", filename, err)
	}
	defer file.Close()

	envVarMap := types.EnvVarMap{}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		key, value := parseLine(line, lineNum, config, &errors, &warnings)

		if key == "" {
			continue
		}

		// this part Checks for duplicates
		if _, exists := envVarMap[key]; exists {
			issue := types.Issue{
				LineNum:   lineNum,
				IssueType: "duplicate",
				Message:   fmt.Sprintf("Duplicate key detected; first occurrence on line %d", envVarMap[key].LineNum),
				KeyName:   key,
			}
			if config.Strict {
				errors = append(errors, issue)
			} else {
				warnings = append(warnings, issue)
			}
		} else {
			envVarMap[key] = types.EnvVar{Value: value, LineNum: lineNum}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return envVarMap, errors, warnings
}

// welp self explanatory init, it guesses the type based on environment variable value
func GuessValueType(value string) string {
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

func parseLine(line string, lineNum int, cfg types.Config, errors, warnings *[]types.Issue) (string, string) {
	line = strings.TrimSpace(line)

	if line == "" || line[0] == '#' {
		return "", ""
	}

	if !strings.Contains(line, "=") {
		*errors = append(*errors, types.Issue{
			LineNum:   lineNum,
			IssueType: "syntax",
			Message:   "missing '='",
			KeyName:   "",
		})
		return "", ""
	}

	parts := strings.SplitN(line, "=", 2)
	return checkAfterSplit(parts, lineNum, cfg, errors, warnings)
}

func checkAfterSplit(parts []string, lineNum int, cfg types.Config, errors, warnings *[]types.Issue) (string, string) {
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		*errors = append(*errors, types.Issue{
			LineNum:   lineNum,
			IssueType: "syntax",
			Message:   "missing key",
			KeyName:   "",
		})
		return "", ""
	}

	if value == "" {
		*warnings = append(*warnings, types.Issue{
			LineNum:   lineNum,
			IssueType: "warning",
			Message:   "missing value",
			KeyName:   key,
		})
	}

	if isQuoted(value) {
		handleQuotedParsing(key, &value, lineNum, cfg, errors, warnings)
	} else {
		handleUnQuotedParsing(key, &value, lineNum, cfg, errors, warnings)
	}

	if cfg.Strict {
		if valid, _ := StrictKeyRegex(key); !valid {
			*errors = append(*errors, types.Issue{
				LineNum:   lineNum,
				IssueType: "strict",
				Message:   "invalid strict key (must be uppercase with underscores)",
				KeyName:   key,
			})
			return "", ""
		}
	}

	return key, value
}

func handleUnQuotedParsing(key string, v *string, lineNum int, cfg types.Config, errors, warnings *[]types.Issue) {
	value := *v

	// Strip inline comments
	if idx := strings.Index(value, "#"); idx != -1 {
		value = strings.TrimSpace(value[:idx])
	}

	// Check for dangling backslash
	if len(value) > 0 && value[len(value)-1] == '\\' {
		issue := types.Issue{
			LineNum:   lineNum,
			IssueType: "warning",
			Message:   "dangling escape at end of value",
			KeyName:   key,
		}
		if cfg.Strict {
			issue.IssueType = "strict"
			*errors = append(*errors, issue)
		} else {
			*warnings = append(*warnings, issue)
		}
	}

	*v = value
}

func handleQuotedParsing(key string, v *string, lineNum int, cfg types.Config, errors, warnings *[]types.Issue) {
	value := *v
	firstChar := value[0]

	for i := 1; i < len(value); i++ {
		if value[i] == firstChar {
			backslashes := 0
			j := i - 1
			for j >= 0 && value[j] == '\\' {
				backslashes++
				j--
			}

			if backslashes%2 == 0 {
				ignoredString := value[i+1:]
				if len(ignoredString) > 0 {
					ignoredString = strings.TrimSpace(ignoredString)
					if len(ignoredString) > 0 && ignoredString[0] != '#' {
						issue := types.Issue{
							LineNum:   lineNum,
							IssueType: "warning",
							Message:   "content after closing quote",
							KeyName:   key,
						}
						if cfg.Strict {
							issue.IssueType = "strict"
							*errors = append(*errors, issue)
						} else {
							*warnings = append(*warnings, issue)
						}
					}
				}

				value = value[1:i]
				*v = value
				return
			}
		} else if value[i] == '\\' {
			if i+1 < len(value) {
				i++
			} else {
				issue := types.Issue{
					LineNum:   lineNum,
					IssueType: "warning",
					Message:   "dangling escape at end of value",
					KeyName:   key,
				}
				if cfg.Strict {
					issue.IssueType = "strict"
					*errors = append(*errors, issue)
				} else {
					*warnings = append(*warnings, issue)
				}
				return
			}
		}

		if i == len(value)-1 {
			issue := types.Issue{
				LineNum:   lineNum,
				IssueType: "warning",
				Message:   "unclosed quoted value",
				KeyName:   key,
			}
			if cfg.Strict {
				issue.IssueType = "strict"
				*errors = append(*errors, issue)
			} else {
				*warnings = append(*warnings, issue)
			}
			return
		}
	}
}

func isQuoted(value string) bool {
	if len(value) >= 2 {
		firstChar := value[0]
		return firstChar == '"' || firstChar == '\''
	}
	return false
}

func StrictKeyRegex(key string) (bool, error) {
	return regexp.MatchString("^[A-Z_][A-Z0-9_]*$", key)
}
