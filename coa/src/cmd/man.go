package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var targetDir string

// manCmd rappresenta il comando per generare la documentazione
var manCmd = &cobra.Command{
	Use:    "man",
	Short:  "Generates man pages for coa",
	Hidden: true, // Importante: non serve all'utente finale, ma solo al builder
	RunE: func(cmd *cobra.Command, args []string) error {
		// Header standard per le man pages Linux
		header := &doc.GenManHeader{
			Title:   "COA",
			Section: "1",
			Source:  "coa " + AppVersion,
			Manual:  "coa User Manual",
		}

		// Assicuriamoci che la cartella di destinazione esista
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		fmt.Printf("Generating man pages in %s...\n", targetDir)

		// Passiamo rootCmd per generare le man pages di tutti i sotto-comandi ricorsivamente
		return doc.GenManTree(rootCmd, header, targetDir)
	},
}

func init() {
	// Definiamo il flag per la directory di output
	manCmd.Flags().StringVarP(&targetDir, "target", "t", "./docs/man", "Target directory for man pages")

	// Aggiungiamo il comando al comando radice (rootCmd)
	rootCmd.AddCommand(manCmd)
}
