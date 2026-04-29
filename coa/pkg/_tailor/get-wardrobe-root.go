package tailor

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func getWardrobeRoot() (string, error) {
	var homeDir string

	// 1. Controlliamo se siamo sotto sudo
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		// Recuperiamo le informazioni dell'utente originale
		u, err := user.Lookup(sudoUser)
		if err == nil {
			homeDir = u.HomeDir
		}
	}

	// 2. Se non siamo sotto sudo o se il lookup è fallito,
	// usiamo la home dell'utente corrente
	if homeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("impossibile determinare la home directory: %v", err)
		}
		homeDir = home
	}

	return filepath.Join(homeDir, ".oa-wardrobe"), nil
}
