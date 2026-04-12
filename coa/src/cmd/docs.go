package cmd

import (
	"coa/src/internal/engine"
	"fmt"

	"github.com/spf13/cobra"
)

var targetDir string

var docsCmd = &cobra.Command{
	Use:    "_gen_docs",
	Short:  "Generate man pages, markdown wiki, and completion scripts",
	Hidden: true, // Completamente invisibile all'utente finale
	RunE: func(cmd *cobra.Command, args []string) error {
		CheckSudoRequirements(cmd.Name(), false)

		if err := engine.HandleDocs(rootCmd, targetDir); err != nil {
			return fmt.Errorf("failed to generate documentation: %w", err)
		}
		return nil
	},
}

func init() {
	// Impostiamo un percorso di default comodo per lo sviluppo locale
	docsCmd.Flags().StringVarP(&targetDir, "target", "t", "./docs", "Base target directory for all documentation types")
	rootCmd.AddCommand(docsCmd)
}
