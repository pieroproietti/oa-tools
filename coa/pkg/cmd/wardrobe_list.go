package cmd

import (
	"coa/pkg/tailor"
	"github.com/spf13/cobra"
)

func wardrobeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Elenca i vestiti disponibili",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Rimosso l'argomento passato: ora è tailor.List()
			return tailor.List()
		},
	}
}
