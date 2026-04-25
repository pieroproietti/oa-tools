package pilot

import (
	"fmt"
	"os"
	"path/filepath"

	"coa/pkg/distro"

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

// DetectAndLoad rileva il sistema e consulta l'index.yaml per trovare lo spartito corretto
func DetectAndLoad() (*Profile, error) {
	// 1. Identità: Chi siamo?
	myDistro := distro.NewDistro()

	indexPath := filepath.Join("coa", "brain.d", "index.yaml")
	//logPilot("Consultazione indice: %s", indexPath)

	// 2. Lettura e parsing dell'indice
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere l'indice %s: %v", indexPath, err)
	}

	var index BrainIndex
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("errore sintassi index.yaml: %v", err)
	}

	// 3. Logica di Matching
	var targetFile string

	//logPilot("Ricerca match per ID='%s' Like=%v", myDistro.DistroID, myDistro.DistroLike)

	for _, entry := range index.Distributions {
		// Match diretto sull'ID (es. "debian" == "debian")
		if entry.ID == myDistro.DistroID {
			targetFile = entry.File
			break
		}

		// Match sulla lista "Like" (es. se siamo su 'ubuntu', troviamo 'debian' perché 'ubuntu' è nei suoi like)
		for _, l := range entry.Like {
			if l == myDistro.DistroID {
				targetFile = entry.File
				break
			}
			// Controllo incrociato sui "Like" della distro rilevata
			/*
				for _, myLike := range myDistro.DistroLike {
					if l == myLike {
						targetFile = entry.File
						break
					}
				}
			*/
			if targetFile != "" {
				break
			}
		}
		if targetFile != "" {
			break
		}
	}

	if targetFile == "" {
		return nil, fmt.Errorf("nessun profilo trovato nell'indice per %s", myDistro.DistroID)
	}

	// 4. Caricamento del profilo finale
	brainPath := filepath.Join("coa", "brain.d", targetFile)
	//LogCoala("Match trovato! Caricamento spartito: %s", brainPath)

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
