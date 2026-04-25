package cmd

import (
	"fmt"

	"coa/pkg/distro"

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
		// Controllo sudo: è un comando informativo, non serve root
		CheckSudoRequirements(cmd.Name(), false)

		// 1. Rileva la distribuzione host
		myDistro := distro.NewDistro()

		// 2. Stampa a video usando i colori centralizzati (Senza passare per l'engine!)
		fmt.Printf("\n%s--- coa distro detect ---%s\n", ColorCyan, ColorReset)
		fmt.Printf("Host Distro:     %s\n", myDistro.DistroID)
		fmt.Printf("Family:          %s\n", myDistro.FamilyID)
		fmt.Printf("DistroLike:      %s\n", myDistro.DistroLike)
		fmt.Printf("Codename:        %s\n", myDistro.CodenameID)
		fmt.Printf("Release:         %s\n", myDistro.ReleaseID)
		fmt.Printf("DistroUniqueID:  %s\n\n", myDistro.DistroUniqueID)
	},
}

func init() {
	rootCmd.AddCommand(detectCmd)
}
