package tailor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Get(repoURL string) error {
	targetDir, err := getWardrobeRoot()
	if err != nil {
		return err
	}

	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
		fmt.Println("♻️  Il wardrobe esiste già. Aggiornamento in corso...")
		return exec.Command("git", "-C", targetDir, "pull").Run()
	}

	os.MkdirAll(targetDir, 0755)
	fmt.Println("📥 Clonazione del nuovo wardrobe...")
	return exec.Command("git", "clone", repoURL, targetDir).Run()
}

// Show ora accetta (costumeName, asJSON) per combaciare con wardrobe.go
func Show(costumeName string, asJSON bool) error {
	root, err := getWardrobeRoot()
	if err != nil {
		return err
	}

	costumePath := filepath.Join(root, "costumes", costumeName, "wardrobe.yaml")
	fmt.Printf("📖 Analisi cartamodello: %s\n", costumePath)

	// Se asJSON è true, qui andrà la logica per stampare la struct in JSON
	if asJSON {
		fmt.Println("[Modalità JSON attivata]")
	}

	return nil
}
