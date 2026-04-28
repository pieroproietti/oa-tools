package pilot

import (
	"fmt"
	"os"
	"path/filepath"

	"coa/pkg/distro"
	"coa/pkg/utils" // Assumendo che i colori siano qui

	"gopkg.in/yaml.v3"
)

// Strutture per il mapping dell'indice
type BrainIndex struct {
	Distributions []DistroMap `yaml:"distributions"`
}

type DistroMap struct {
	ID   string   `yaml:"id"`
	Like []string `yaml:"like"`
	File string   `yaml:"file"`
}

// DetectAndLoad rileva il sistema e consulta l'index.yaml per trovare lo spartito corretto.
// Implementa il fallback tra ambiente di sviluppo locale e directory di sistema /etc.
func DetectAndLoad() (*Profile, error) {
	// 1. Identità: Chi siamo?
	myDistro := distro.NewDistro()
	

	// 2. Ricerca del percorso della configurazione (Dev vs System)
	var baseDir string
	pathsToTry := []string{
		filepath.Join("coa", "brain.d"),          // Percorso di Sviluppo
		"/etc/oa-tools.d/brain.d",                // Percorso di Produzione
	}

	for _, path := range pathsToTry {
		if _, err := os.Stat(filepath.Join(path, "index.yaml")); err == nil {
			baseDir = path
			break
		}
	}

	if baseDir == "" {
		return nil, fmt.Errorf("nessuna configurazione brain trovata nei percorsi previsti")
	}

	indexPath := filepath.Join(baseDir, "index.yaml")
	utils.LogCoala("%s[pilot]%s Utilizzo indice: %s", utils.ColorCyan, utils.ColorReset, indexPath)

	// 3. Lettura e parsing dell'indice
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere l'indice %s: %v", indexPath, err)
	}

	var index BrainIndex
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("errore sintassi index.yaml: %v", err)
	}

	// 4. Logica di Matching: Cerchiamo il profilo adatto alla distro corrente
	var targetFile string
	for _, entry := range index.Distributions {
		// Match diretto sull'ID (es. "debian" == "debian")
		if entry.ID == myDistro.DistroID {
			targetFile = entry.File
			break
		}

		// Match sulla lista "Like"
		for _, l := range entry.Like {
			if l == myDistro.DistroID {
				targetFile = entry.File
				break
			}
		}
		if targetFile != "" {
			break
		}
	}

	if targetFile == "" {
		return nil, fmt.Errorf("nessun profilo trovato nell'indice per %s (ID: %s)", myDistro.DistroLike, myDistro.DistroID)
	}

	// 5. Caricamento del profilo finale
	// Usiamo la stessa baseDir dove abbiamo trovato l'index.yaml
	brainPath := filepath.Join(baseDir, targetFile)
	utils.LogCoala("%s[pilot]%s Caricamento profilo: %s", utils.ColorCyan, utils.ColorReset, targetFile)

	data, err := os.ReadFile(brainPath)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere il profilo %s: %v", brainPath, err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("errore sintassi nel profilo %s: %v", brainPath, err)
	}

	return &profile, nil
}