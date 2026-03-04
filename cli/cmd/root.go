package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kr",
	Short: "A CLI tool for managing Krypton",
	Long:  `kr is a command-line interface for managing and interacting with the Krypton server.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(loginCmd())
}
