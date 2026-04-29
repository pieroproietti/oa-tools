package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"os"
)

func Get() error {
	root, err := getWardrobeRoot()
	if err != nil { return err }

	if _, err := os.Stat(root); os.IsNotExist(err) {
		utils.LogCoala("Scaricamento di oa-wardrobe...")
		cmd := fmt.Sprintf("git clone https://github.com/pieroproietti/oa-wardrobe.git %s", root)
		return utils.Exec(cmd)
	}
	
	utils.LogCoala("oa-wardrobe è già presente in %s. Per aggiornare usa git pull manualmente.", root)
	return nil
}