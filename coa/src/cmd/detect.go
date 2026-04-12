package cmd

import (
	"github.com/spf13/cobra"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Display detected host system information",
	Long: `The 'detect' command is a read-only diagnostic utility for the user. 

It performs a quick scan of the host environment to identify the running GNU/Linux distribution, its parent family (e.g., mapping Ubuntu to the Debian family), and the hardware architecture. 

It does not save this state or alter any configuration; it simply provides a clear overview of the environment 'coa' is currently running in.`,
	Example: `  # Display the host system profile
  coa detect`,
	Run: func(cmd *cobra.Command, args []string) {
		// Logica di visualizzazione (es. GetDiscovery() e print)
	},
}

func init() {
	rootCmd.AddCommand(detectCmd)
}
