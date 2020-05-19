package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version.",
	Long:  `Prints the current version of the tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nova version " + currentVersion)
	},
}
