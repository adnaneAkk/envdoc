package types

// Config struct for CLI options
type Config struct {
	Strict         bool
	GenerateSchema bool
	Unmask         bool
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
	Value     string `yaml:"example"`
	Type      string `yaml:"type"`
	Required  bool   `yaml:"required"`
	Sensitive bool   `yaml:"sensitive"`
}
type Schema map[string]SchemaItem

type Diff struct {
	DiffType string
	Message  string
	KeyName  string //the relevant key that might be missing ?? not sure
	Value1   string //this is incase there is a different value between two files with the same key
	Value2   string
}

type DiffMap []Diff
