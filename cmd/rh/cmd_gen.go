package main

import (
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Code generation commands",
}

func init() {
	genCmd.AddCommand(genDaoCmd)
	genCmd.AddCommand(genServiceCmd)
}
