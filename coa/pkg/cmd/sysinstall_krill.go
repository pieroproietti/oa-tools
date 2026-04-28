package cmd

import (
	"coa/pkg/utils"
	"os"
)

// runKrillInstaller è il segnaposto per l'installatore testuale.
// Per ora non rompe i coglioni e ci permette di compilare tutto.
func runKrillInstaller() {
	utils.LogCoala("%s[Krill]%s L'installatore TUI non è ancora pronto. Usa Calamares per ora!", utils.ColorYellow, utils.ColorReset)
	
	// Usciamo puliti
	os.Exit(0)
}