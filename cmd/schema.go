package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/adnaneAkk/envdoc/internal/parser"
	"github.com/adnaneAkk/envdoc/internal/schema"
	"github.com/adnaneAkk/envdoc/internal/types"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	outputFile   string
)

var schemaCmd = &cobra.Command{
	Use:   "schema [.env file]",
	Short: "Generate schema from .env file",
	Long:  `Generate a JSON or YAML schema documenting all environment variables`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			envFile = args[0]
		} else {
			envFile = ".env"
		}
		unmask, _ := cmd.Flags().GetBool("unmask")
		if unmask {
			fmt.Fprint(os.Stderr, "⚠  This will expose sensitive values in output. Continue? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Fprintln(os.Stderr, "Aborted.")
				os.Exit(0)
			}
		}
		runSchemaGeneration(envFile, strict, unmask, outputFormat, outputFile)
	},
}

func init() {
	schemaCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json|yaml|text)")
	schemaCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	schemaCmd.Flags().Bool("unmask", false, "unmask sensitive values in output")
	rootCmd.AddCommand(schemaCmd)
}

func runSchemaGeneration(filename string, strictMode, unmask bool, format string, outFile string) {
	config := types.Config{
		Strict: strictMode,
		Unmask: unmask,
	}

	// Parse the file
	envVarMap, errors, warnings, err := parser.ParseFile(filename, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Print errors/warnings
	if len(errors) > 0 {
		fmt.Printf("Errors: %d found\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  Line %d [%s]: %s (Key: %s)\n", e.LineNum, e.IssueType, e.Message, e.KeyName)
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %d found\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("  Line %d [%s]: %s (Key: %s)\n", w.LineNum, w.IssueType, w.Message, w.KeyName)
		}
	}

	// Generate schema
	schemaData := schema.Generate(envVarMap, config)

	// Output based on format
	output, err := schema.Output(schemaData, format)
	if err != nil {
		log.Fatalf("Error generating output: %v", err)
	}

	// Write to file or stdout
	if outFile != "" {
		if err := os.WriteFile(outFile, []byte(output), 0644); err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
		fmt.Printf("\n✓ Schema written to %s\n", outFile)
	} else {
		fmt.Println(output)
	}

	// Exit with error if errors found
	if len(errors) > 0 {
		os.Exit(1)
	}
}
