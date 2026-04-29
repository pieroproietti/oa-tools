package cmd

import (
	"coa/pkg/tailor"
	"github.com/spf13/cobra"
)

func wardrobeShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [costume]",
		Short: "Mostra i dettagli di un vestito",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ora accetta solo il nome del costume
			return tailor.Show(args[0])
		},
	}
}