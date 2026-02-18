package cmd

import (
	"fmt"
	"os"

	"github.com/adnaneAkk/envdoc/internal/parser"
	"github.com/adnaneAkk/envdoc/internal/types"

	"github.com/spf13/cobra"
)

var (
	strict  bool
	envFile string
)

var rootCmd = &cobra.Command{
	Use:   "envdoc [.env file]",
	Short: "Parse and validate .env files",
	Long:  `A fast and flexible .env file parser with schema generation and validation`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			envFile = args[0]
		} else {
			envFile = ".env"
		}
		runValidation(envFile, strict)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&strict, "strict", "s", false, "Enable strict mode")
}

func Execute() error {
	return rootCmd.Execute()
}

func runValidation(filename string, strictMode bool) {
	config := types.Config{
		Strict: strictMode,
	}

	envVarMap, errors, warnings, err := parser.ParseFile(filename, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Print errors
	if len(errors) > 0 {
		fmt.Printf("Errors: %d found\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  Line %d [%s]: %s (Key: %s)\n", e.LineNum, e.IssueType, e.Message, e.KeyName)
		}
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Printf("\nWarnings: %d found\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("  Line %d [%s]: %s (Key: %s)\n", w.LineNum, w.IssueType, w.Message, w.KeyName)
		}
	}

	// Success message
	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Printf("\n✓ %s is valid! Found %d environment variables.\n", filename, len(envVarMap))
	}

	// Exit with error code if errors found
	if len(errors) > 0 {
		os.Exit(1)
	}
}
