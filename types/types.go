package types

type Config struct {
	Strict bool
}

type Issue struct {
	lineNum   int    `records line number`
	IssueType string `records the issue type : syntax, duplicate, strict, warning`
	Message   string `Issue message`
	KeyName   string ``
}
