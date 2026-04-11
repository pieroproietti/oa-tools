package assets

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed configs/*
var internalConfigs embed.FS

// ExtractConfigs estrae le configurazioni incorporate nel binario verso una directory temporanea
func ExtractConfigs(destRoot string) error {
	fmt.Printf("[DEBUG] Tentativo di estrazione in: %s\n", destRoot)

	// Creiamo la radice
	if err := os.MkdirAll(destRoot, 0755); err != nil {
		return err
	}

	// Vediamo cosa c'è nell'embed prima di copiare
	entries, err := internalConfigs.ReadDir("configs")
	if err != nil {
		return fmt.Errorf("L'EMBED È VUOTO O ERRATO: %w", err)
	}
	fmt.Printf("[DEBUG] Trovate %d entries nella cartella embed 'configs'\n", len(entries))

	return fsCopy(internalConfigs, "configs", destRoot)
}

func fsCopy(fs embed.FS, src, dest string) error {
	entries, err := fs.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			if err := fsCopy(fs, srcPath, destPath); err != nil {
				return err
			}
		} else {
			data, err := fs.ReadFile(srcPath)
			if err != nil {
				return err
			}
			// Assicuriamoci che la directory padre esista (es. configs/mkinitcpio)
			os.MkdirAll(filepath.Dir(destPath), 0755)
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
