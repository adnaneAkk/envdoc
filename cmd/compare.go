package cmd

import (
	"fmt"
	"os"

	"github.com/adnaneAkk/envdoc/internal/parser"
	"github.com/adnaneAkk/envdoc/internal/types"
	"github.com/spf13/cobra"
)

var (
	envFile1 string
	envFile2 string
)

var compareCmd = &cobra.Command{

	Use:   "compare [.env file1 ] [.env file2]",
	Short: "Compares the second .env file to the first one",
	Long:  `Compares two env files and report them to see the differences between them`,
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var e1, e2 string
		if len(args) >= 1 {
			e1 = args[0]
		} else {
			e1 = envFile1
		}

		if len(args) == 2 {
			e2 = args[1]
		} else {
			e2 = envFile2
		}

		if e1 == "" || e2 == "" {
			fmt.Println("Error: you must provide two env files (either as args or flags)")
			cmd.Usage()
			os.Exit(1)
		}

		runCompare(e1, e2, strict)
	},
}

func init() {
	compareCmd.Flags().StringVar(&envFile1, "env1", "", "First env file")
	compareCmd.Flags().StringVar(&envFile2, "env2", "", "Second env file")

	rootCmd.AddCommand(compareCmd)
}

func runCompare(envfile1, envfile2 string, strictMode bool) {
	config := types.Config{
		Strict: strictMode,
	}

	EnvMap1, File1erors, File1warnings, err := parser.ParseFile(envfile1, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	EnvMap2, File2erors, File2warnings, err := parser.ParseFile(envfile2, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(File1erors) > 0 || len(File1warnings) > 0 {
		fmt.Printf("\n=== Parse Issues: %s ===\n", envfile1)
		for _, e := range File1erors {
			fmt.Printf("  ✗ Line %d [%s]: %s (Key: %s)\n", e.LineNum, e.IssueType, e.Message, e.KeyName)
		}
		for _, w := range File1warnings {
			fmt.Printf("  ⚠ Line %d [%s]: %s (Key: %s)\n", w.LineNum, w.IssueType, w.Message, w.KeyName)
		}
	}

	// Parse issues for file2
	if len(File2erors) > 0 || len(File2warnings) > 0 {
		fmt.Printf("\n=== Parse Issues: %s ===\n", envfile2)
		for _, e := range File2erors {
			fmt.Printf("  ✗ Line %d [%s]: %s (Key: %s)\n", e.LineNum, e.IssueType, e.Message, e.KeyName)
		}
		for _, w := range File2warnings {
			fmt.Printf("  ⚠ Line %d [%s]: %s (Key: %s)\n", w.LineNum, w.IssueType, w.Message, w.KeyName)
		}
	}

	//this is going to be responsable for the difference reporting
	var difference types.DiffMap

	for key1, value1 := range EnvMap1 {
		value2, exists := EnvMap2[key1]

		if !exists {
			difference = append(difference, types.Diff{
				DiffType: "missing key",
				Message:  fmt.Sprintf("file %s is missing the key %s file %s has (line %d)", envFile2, key1, envFile1, value1.LineNum),
				KeyName:  key1,
				Value1:   value1.Value,
				Value2:   "",
			},
			)
		} else if value2.Value == value1.Value {
			continue
		} else {
			difference = append(difference, types.Diff{
				DiffType: "difference in value",
				Message:  fmt.Sprintf(" key %s has value %s in file %s (line %d) ,but %s in file %s (line %d)", key1, value1.Value, envFile1, value1.LineNum, value2.Value, envFile2, value2.LineNum),
				KeyName:  key1,
				Value1:   value1.Value,
				Value2:   value2.Value,
			},
			)
		}
	}

	for key2, value2 := range EnvMap2 {
		_, exists := EnvMap1[key2]
		if !exists {
			difference = append(difference, types.Diff{
				DiffType: "missing key",
				Message:  fmt.Sprintf("file %s is missing the key %s file %s has (line %d)", envFile1, key2, envFile2, value2.LineNum),
				KeyName:  key2,
				Value1:   "",
				Value2:   value2.Value,
			},
			)
		}
	}

	if len(difference) == 0 {
		fmt.Println("\n✓ Files are identical")
	} else {
		fmt.Printf("\n=== Comparison: %s vs %s ===\n", envfile1, envfile2)
		for _, d := range difference {
			switch d.DiffType {
			case "missing key":
				if d.Value1 != "" {
					fmt.Printf("  - %-20s (only in %s)\n", d.KeyName, envfile1)
				} else {
					fmt.Printf("  + %-20s (only in %s)\n", d.KeyName, envfile2)
				}
			case "difference in value":
				fmt.Printf("  ~ %-20s %q → %q\n", d.KeyName, d.Value1, d.Value2)
			}
		}
		fmt.Printf("\n%d difference(s) found\n", len(difference))
	}

}
