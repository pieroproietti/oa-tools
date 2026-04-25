package cmd

import (
	"coa/pkg/builder"
	"coa/pkg/distro"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile binaries and generate native distribution packages (.deb, PKGBUILD)",
	Long: `The 'build' command is the integrated packaging tool for the coa/oa ecosystem.
It orchestrates the full compilation of both the C-native engine (oa) and the Go-based orchestrator (coa), triggers the automatic generation of documentation and shell completions, and finally packages everything into native distribution formats like .deb (Debian/Ubuntu) or PKGBUILD (Arch Linux).`,
	Example: `  # Compile the ecosystem and generate native packages
  coa build`,
	Run: func(cmd *cobra.Command, args []string) {
		// Non richiede i privilegi di root
		CheckSudoRequirements(cmd.Name(), false)

		// Rileva la distribuzione host (i Sensi)
		myDistro := distro.NewDistro()

		// Passa la palla al motore di build, includendo la versione di Git
		builder.HandleBuild(myDistro, AppVersion)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
