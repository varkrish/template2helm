package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "t2hc",
		Short: "t2hc converts an OpenShift Template into a Helm Chart inline with Helm Common Chart.",
		Long: `t2hc converts an OpenShift Template into a Helm Chart inline with Helm Common Chart.
      For more info, check out https://github.com/varkrish/template2helm`,
	}
)

// Execute - entrypoint for CLI tool
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
