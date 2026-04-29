package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"path/filepath"
)

func Show(costumeName string) error {
	root, err := getWardrobeRoot()
	if err != nil { return err }

	costumeDir := filepath.Join(root, "costumes", costumeName)
	yamlPath := findYaml(costumeDir)
	suit, err := loadSuit(yamlPath)
	if err != nil { return err }

	fmt.Printf(utils.ColorCyan+"Costume: %s\n"+utils.ColorReset, suit.Name)
	fmt.Printf("Descrizione: %s\n", suit.Description)
	fmt.Printf("Pacchetti: %v\n", suit.Packages)
	return nil
}
