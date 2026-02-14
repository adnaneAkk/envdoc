package types

// Config struct for CLI options
type Config struct {
	Strict         bool
	GenerateSchema bool
}

// Issue struct for recording issues found in .env
type Issue struct {
	LineNum   int
	IssueType string // for now there is : syntax, duplicate, strict, warning
	Message   string
	KeyName   string
}

// this is gonna be for the final report
type Report struct {
	Errors   []Issue
	Warnings []Issue
}

// refers to environment variable
type EnvVar struct {
	Value   string
	LineNum int
}
type EnvVarMap map[string]EnvVar

type SchemaItem struct {
	Example  string `yaml:"example"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
}
type Schema map[string]SchemaItem
