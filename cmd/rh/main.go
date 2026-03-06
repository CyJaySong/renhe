package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rh",
	Short: "RenHe framework CLI tool",
	Long:  "rh is a CLI tool for project scaffolding and code generation based on the RenHe framework.",
}

func main() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
