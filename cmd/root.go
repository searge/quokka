// Package cmd implements the CLI for wombat.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "wombat",
	Short: "A resilient software forge platform",
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help()
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}
