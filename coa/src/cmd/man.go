package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var targetDir string

// manCmd rappresenta il comando per generare la documentazione
var manCmd = &cobra.Command{
	Use:    "man",
	Short:  "Generates man and markdown documentation for coa",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		// 1. Percorsi di destinazione
		manPath := filepath.Join(targetDir, "man")
		mdPath := filepath.Join(targetDir, "md")

		// Creiamo le directory
		if err := os.MkdirAll(manPath, 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(mdPath, 0755); err != nil {
			return err
		}

		// 2. Generazione MAN PAGES
		header := &doc.GenManHeader{
			Title:   "COA",
			Section: "1",
			Source:  "coa " + AppVersion,
			Manual:  "coa User Manual",
		}
		fmt.Printf("Generating man pages in %s...\n", manPath)
		if err := doc.GenManTree(rootCmd, header, manPath); err != nil {
			return fmt.Errorf("failed to generate man pages: %w", err)
		}

		// 3. Generazione MARKDOWN
		fmt.Printf("Generating markdown docs in %s...\n", mdPath)
		if err := doc.GenMarkdownTree(rootCmd, mdPath); err != nil {
			return fmt.Errorf("failed to generate markdown docs: %w", err)
		}

		fmt.Println("Documentation generated successfully.")
		return nil
	},
}

func init() {
	// Definiamo il flag per la directory base di output
	// Di default ora punta a ./docs e creerà ./docs/man e ./docs/md

	manCmd.Flags().StringVarP(&targetDir, "target", "t", "./coa/docs", "Base target directory for documentation")
	rootCmd.AddCommand(manCmd)
}
