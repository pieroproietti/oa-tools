package cmd

import (
	"coa/pkg/tailor"
	"github.com/spf13/cobra"
)

func wardrobeWearCmd() *cobra.Command {
	var noAcc bool
	var noFirm bool

	cmd := &cobra.Command{
		Use:   "wear [costume]",
		Short: "Indossa un vestito (costume) dal guardaroba",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Passiamo le flags alla funzione tailor.Wear
			return tailor.Wear(args[0], noAcc, noFirm)
		},
	}

	// Definiamo le flags qui
	cmd.Flags().BoolVar(&noAcc, "no-acc", false, "Non installare gli accessori")
	cmd.Flags().BoolVar(&noFirm, "no-firm", false, "Non installare il firmware")

	return cmd
}