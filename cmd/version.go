package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  `Prints the version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v1.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
