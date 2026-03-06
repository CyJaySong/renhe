package main

import (
	"fmt"

	"github.com/cyjaysong/renhe/cmd/rh/internal/config"
	"github.com/cyjaysong/renhe/cmd/rh/internal/gen"
	"github.com/spf13/cobra"
)

var genServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generate service interfaces from logic layer code",
	Long: `Generate service interfaces by parsing Go source files in the logic directory.

Configuration is read from hack/config.yaml:

  rh:
    gen:
      service:
        srcPath: "internal/logic"      # logic source directory (default)
        dstPath: "internal/service"    # service output directory (default)

The tool scans each sub-package under srcPath, extracts exported methods
from structs, and generates corresponding interface files in dstPath.`,
	RunE: runGenService,
}

func runGenService(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("hack/config.yaml")
	if err != nil {
		return err
	}

	module, err := config.DetectModule(".")
	if err != nil {
		return fmt.Errorf("failed to detect go module: %w", err)
	}

	sc := cfg.RH.Gen.Service
	genCfg := gen.ServiceConfig{
		SrcPath: sc.SrcPath,
		DstPath: sc.DstPath,
		Module:  module,
	}

	return gen.RunService(genCfg)
}
