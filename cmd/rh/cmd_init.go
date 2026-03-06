package main

import (
	initcmd "github.com/cyjaysong/renhe/cmd/rh/internal/init"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name] [relative-path]",
	Short: "Initialize a new RenHe project with standard directory structure",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return initcmd.Run(args[0], args)
	},
}
