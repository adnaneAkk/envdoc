package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/adnaneAkk/envdoc/internal/types"
)

// ParseFile parses an .env file and returns the parsed map, errors, and warnings
func ParseFile(filename string, config types.Config) (types.EnvVarMap, []types.Issue, []types.Issue) {
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

		// Check for duplicates
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

func parseLine(line string, lineNum int, cfg types.Config, errors, warnings *[]types.Issue) (string, string) {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || line[0] == '#' {
		return "", ""
	}

	// Check for equals sign
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

	// Validate key
	if key == "" {
		*errors = append(*errors, types.Issue{
			LineNum:   lineNum,
			IssueType: "syntax",
			Message:   "missing key",
			KeyName:   "",
		})
		return "", ""
	}

	// Warn if value is empty
	if value == "" {
		*warnings = append(*warnings, types.Issue{
			LineNum:   lineNum,
			IssueType: "warning",
			Message:   "missing value",
			KeyName:   key,
		})
	}

	// Handle quoted or unquoted values
	if isQuoted(value) {
		handleQuotedParsing(key, &value, lineNum, cfg, errors, warnings)
	} else {
		handleUnQuotedParsing(key, &value, lineNum, cfg, errors, warnings)
	}

	// Strict mode key validation
	if cfg.Strict {
		if valid, _ := strictKeyRegex(key); !valid {
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
			// Count preceding backslashes
			backslashes := 0
			j := i - 1
			for j >= 0 && value[j] == '\\' {
				backslashes++
				j--
			}

			// Even number = not escaped
			if backslashes%2 == 0 {
				// Check for content after closing quote
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
			// Skip escaped character
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

		// Unclosed quote
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

func strictKeyRegex(key string) (bool, error) {
	return regexp.MatchString("^[A-Z_][A-Z0-9_]*$", key)
}
