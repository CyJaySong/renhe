package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const Version = "v0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of rh CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("RenHe CLI %s\n", Version)

		fmt.Println("Env Detail:")
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		fmt.Println("CLI Detail:")
		if execPath, err := os.Executable(); err == nil {
			fmt.Printf("  Installed At: %s\n", execPath)
		}
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("  Built Go Version: %s\n", info.GoVersion)
		}
	},
}
