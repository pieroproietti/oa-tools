package engine

import (
	"fmt"
	"os"
	"os/exec"
)

// HandleKill gestisce la pulizia profonda invocando prima oa e poi rimuovendo la directory
func HandleKill() {
	fmt.Println("\033[1;33m[coa]\033[0m Freeing the nest...")
	oaPath := getOaPath() // Questa funzione l'avevamo già messa in produce.go
	cmd := exec.Command("sudo", oaPath, "cleanup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Cleanup (unmount) failed: %v\n", err)
	}

	workPath := "/home/eggs"
	fmt.Printf("\033[1;31m[coa]\033[0m Removing workspace: %s\n", workPath)
	rmCmd := exec.Command("sudo", "rm", "-rf", workPath)
	if err := rmCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Physical removal failed: %v\n", err)
	} else {
		fmt.Println("\033[1;32m[coa]\033[0m Nest is empty. System clean.")
	}
}
