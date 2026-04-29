package tailor

import (
	"coa/pkg/distro" // Importante: serve per il rilevamento
	"coa/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func List(distroFilter string) error {
	root, err := getWardrobeRoot()
	if err != nil {
		return err
	}

	utils.LogCoala("Esplorazione wardrobe in: %s", root)

	// Recuperiamo la distro attuale per il confronto dell'asterisco
	d := distro.NewDistro()
	currentDistro := strings.ToLower(d.DistroLike)

	categories := []string{"costumes", "accessories"}

	for _, cat := range categories {
		catPath := filepath.Join(root, cat)
		entries, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}

		fmt.Printf("\n--- %s ---\n", strings.ToUpper(cat))
		for _, e := range entries {
			if e.IsDir() {
				costumePath := filepath.Join(catPath, e.Name())

				// 1. Usiamo la logica di findCompatibleYaml (definita in wear_logic.go)
				yamlPath := findCompatibleYaml(costumePath)

				// 2. Fallback se non trova nulla di specifico
				if yamlPath == "" {
					yamlPath = findFirstYaml(costumePath)
				}

				if yamlPath != "" {
					if data, err := os.ReadFile(yamlPath); err == nil {
						var suit Suit
						if err := yaml.Unmarshal(data, &suit); err == nil {

							// Filtro distro (es: -d debian)
							if distroFilter == "" || strings.Contains(strings.ToLower(suit.Distro), strings.ToLower(distroFilter)) {
								status := ""

								// Se il file caricato non contiene il nome della nostra distro, mettiamo l'asterisco
								if !strings.Contains(strings.ToLower(yamlPath), currentDistro) {
									status = utils.ColorRed + "*" + utils.ColorReset
								}

								fmt.Printf("  📦 %-15s | %s %s\n", e.Name(), suit.Description, status)
							}
						}
					}
				} else {
					fmt.Printf("  📁 %-15s | (Nessun file .yaml trovato)\n", e.Name())
				}
			}
		}
	}
	return nil
}

func findFirstYaml(dir string) string {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") {
			return filepath.Join(dir, f.Name())
		}
	}
	return ""
}
