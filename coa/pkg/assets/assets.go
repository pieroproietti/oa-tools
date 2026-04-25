package assets

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

// --- SISTEMA DI LOGGING ASSETS ---
const (
	ColorCyan  = "\033[1;36m"
	ColorReset = "\033[0m"
)

func logAssets(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[coa-assets]%s %s\n", ColorCyan, ColorReset, msg)
}

// ---------------------------------

//go:embed configs/*
var internalConfigs embed.FS

//go:embed calamares_base/*
var calamaresFiles embed.FS

// ExtractConfigs estrae le configurazioni incorporate nel binario verso una directory temporanea
func ExtractConfigs(destRoot string) error {
	logAssets("Estrazione configurazioni base in: %s", destRoot)

	// Creiamo la radice
	if err := os.MkdirAll(destRoot, 0755); err != nil {
		return fmt.Errorf("impossibile creare la directory %s: %v", destRoot, err)
	}

	// Verifica integrità dell'embed
	entries, err := internalConfigs.ReadDir("configs")
	if err != nil {
		return fmt.Errorf("l'embed 'configs' è vuoto o errato: %w", err)
	}

	logAssets("Trovati %d elementi nella cartella embed 'configs'", len(entries))

	return fsCopy(internalConfigs, "configs", destRoot)
}

// ExtractCalamares estrae i file universali di Calamares usando fsCopy
func ExtractCalamares(destRoot string) error {
	logAssets("Estrazione asset Calamares in: %s", destRoot)

	if err := os.MkdirAll(destRoot, 0755); err != nil {
		return fmt.Errorf("impossibile creare la directory %s: %v", destRoot, err)
	}

	return fsCopy(calamaresFiles, "calamares_base", destRoot)
}

// fsCopy copia ricorsivamente i file da un filesystem virtuale embed a quello fisico
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
