package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Free the nest and unmount filesystems",
	Long: `Safely tears down the remastering environment. 
It uses MNT_DETACH to unmount the OverlayFS and virtual API filesystems (/dev, /proc, /sys) without affecting the running host, then removes the temporary workspace.`,
	Example: `  # Clean up the default workspace
  sudo coa kill`,
	Run: func(cmd *cobra.Command, args []string) {
		// Controllo sudo: smontare filesystem e cancellare /home/eggs richiede i privilegi
		CheckSudoRequirements(cmd.Name(), true)
		handleKill()
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
}

// =====================================================================
// LOGICA DI PULIZIA (Ex-Engine)
// =====================================================================

// handleKill gestisce la pulizia profonda invocando prima oa e poi rimuovendo la directory
func handleKill() {
	LogCoala("Freeing the nest...")

	// 1. Chiamiamo il motore C per smontare in sicurezza i mountpoint
	// Nota: assumiamo che 'oa' sia nel PATH, come in remaster.go
	cmd := exec.Command("oa", "cleanup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		LogError("Cleanup (unmount) failed: %v", err)
		// Non blocchiamo l'esecuzione qui, proviamo comunque a rimuovere la cartella
	}

	// 2. Rimozione fisica della workspace
	workPath := "/home/eggs"
	LogCoala("Removing workspace: %s", workPath)

	rmCmd := exec.Command("rm", "-rf", workPath)
	rmCmd.Stdout = os.Stdout
	rmCmd.Stderr = os.Stderr

	if err := rmCmd.Run(); err != nil {
		LogError("Physical removal failed: %v", err)
	} else {
		LogSuccess("Nest is empty. System clean.")
	}
}
