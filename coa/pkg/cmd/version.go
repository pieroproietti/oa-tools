package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of coa",
	Run: func(cmd *cobra.Command, args []string) {
		// Non richiede permessi di root
		CheckSudoRequirements(cmd.Name(), false)
		fmt.Printf("coa %s - The Mind of remaster\n", AppVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
