package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/hyt4/envcheck/internal/parser"
	"github.com/hyt4/envcheck/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	envFile     string
	exampleFile string
	dir         string
	ci          bool
	format      string
)

var rootCmd = &cobra.Command{
	Use:   "envcheck",
	Short: "Check your .env file for missing, undocumented, and unused variables",
	Long: `envcheck scans your project and reports:
  - Missing variables (in .env.example but not in .env)
  - Undocumented variables (in .env but not in .env.example)
  - Unused variables (defined but never referenced in source code)`,
	Run: runCheck,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&envFile, "env", "e", ".env", "path to .env file")
	rootCmd.Flags().StringVarP(&exampleFile, "example", "x", ".env.example", "path to .env.example file")
	rootCmd.Flags().StringVarP(&dir, "dir", "d", ".", "project root directory to scan")
	rootCmd.Flags().BoolVar(&ci, "ci", false, "output in CI-friendly format")
	rootCmd.Flags().StringVarP(&format, "format", "f", "text", "output format (text or json)")
}

func runCheck(cmd *cobra.Command, args []string) {
	bold := color.New(color.Bold)

	actual, err := parser.ParseFile(envFile)
	if err != nil {
		color.Red("Error reading %s: %v", envFile, err)
		os.Exit(1)
	}

	example, err := parser.ParseFile(exampleFile)
	if err != nil {
		color.Red("Error reading %s: %v", exampleFile, err)
		os.Exit(1)
	}

	missing, undocumented := parser.Diff(actual, example)

	usedKeys, err := scanner.FindUsedKeys(dir)
	if err != nil {
		color.Red("Error scanning directory: %v", err)
		os.Exit(1)
	}
	unused := scanner.FindUnused(actual, usedKeys)

	hasIssues := len(missing) > 0 || len(undocumented) > 0 || len(unused) > 0

	if format == "json" {
		printJSON(missing, undocumented, unused)
	} else {
		printText(bold, missing, undocumented, unused, hasIssues)
	}

	if ci && hasIssues {
		os.Exit(1)
	}
}

func printText(bold *color.Color, missing, undocumented, unused []string, hasIssues bool) {
	if len(missing) > 0 {
		bold.Println("\n Missing (in .env.example but not in .env):")
		for _, key := range missing {
			color.Red("   ✗ %s", key)
		}
	}

	if len(undocumented) > 0 {
		bold.Println("\n Undocumented (in .env but not in .env.example):")
		for _, key := range undocumented {
			color.Yellow("   ⚠ %s", key)
		}
	}

	if len(unused) > 0 {
		bold.Println("\n Unused (defined in .env but never referenced in code):")
		for _, key := range unused {
			color.Magenta("   ~ %s", key)
		}
	}

	if !hasIssues {
		color.Green("\n ✓ All good! Your .env matches .env.example\n")
	} else {
		fmt.Println()
	}
}

func printJSON(missing, undocumented, unused []string) {
	// build simple JSON manually to avoid importing encoding/json
	fmt.Printf(`{"missing":%s,"undocumented":%s,"unused":%s}`+"\n",
		toJSONArray(missing),
		toJSONArray(undocumented),
		toJSONArray(unused),
	)
}

func toJSONArray(keys []string) string {
	if len(keys) == 0 {
		return "[]"
	}
	result := "["
	for i, k := range keys {
		if i > 0 {
			result += ","
		}
		result += `"` + k + `"`
	}
	return result + "]"
}
