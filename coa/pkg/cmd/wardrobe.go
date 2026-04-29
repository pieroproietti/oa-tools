package cmd

import (
	"github.com/spf13/cobra"
)

// Questa funzione viene chiamata automaticamente da Go
func init() {
	// Presumendo che il tuo comando radice si chiami 'rootCmd'
	// Se nel tuo root.go si chiama in modo diverso (es. 'coaCmd'), usa quel nome
	rootCmd.AddCommand(wardrobeCmd())
}

func wardrobeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wardrobe",
		Short: "Gestione dei vestiti (costumes)",
	}

	// Attacchiamo i rami al comando wardrobe
	cmd.AddCommand(wardrobeGetCmd())
	cmd.AddCommand(wardrobeListCmd())
	cmd.AddCommand(wardrobeShowCmd())
	cmd.AddCommand(wardrobeWearCmd())

	return cmd
}