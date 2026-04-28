package cmd

import (
	"coa/pkg/tailor"
	"fmt"

	"github.com/spf13/cobra"
)

const WardrobeDefaultRepo = "https://github.com/pieroproietti/oa-wardrobe.git"
var getCmd = &cobra.Command{
	Use:   "get [REPO_URL]",
	Short: "Scarica un wardrobe da un repository",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		
		repo := WardrobeDefaultRepo
		if len(args) > 0 {
			repo = args[0]
		}

		fmt.Printf("🌐 Recupero wardrobe da: %s\n", repo)
		if err := tailor.Get(repo); err != nil {
			fmt.Println("❌ Errore durante il Get:", err)
		}
	},
}
