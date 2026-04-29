package cmd

import (
	"coa/pkg/tailor"
	"github.com/spf13/cobra"
)

func wardrobeGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Scarica o aggiorna il guardaroba",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tailor.Get()
		},
	}
}