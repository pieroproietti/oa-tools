package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
)

func List() error {
	root, err := getWardrobeRoot()
	if err != nil { return err }

	costumesDir := filepath.Join(root, "costumes")
	entries, err := os.ReadDir(costumesDir)
	if err != nil { return fmt.Errorf("impossibile leggere i costumi: %v", err) }

	fmt.Println(utils.ColorCyan + "Vestiti disponibili in oa-wardrobe:" + utils.ColorReset)
	for _, entry := range entries {
		if entry.IsDir() {
			yamlPath := findYaml(filepath.Join(costumesDir, entry.Name()))
			if yamlPath != "" {
				if suit, err := loadSuit(yamlPath); err == nil {
					fmt.Printf("- %-12s: %s\n", utils.ColorYellow+entry.Name()+utils.ColorReset, suit.Description)
				}
			}
		}
	}
	return nil
}
