package cmd

import (
	"coa/pkg/calamares"
	"coa/pkg/pilot"
	"coa/pkg/utils"
)

// RunCalamaresInstaller coordina la preparazione e il lancio di Calamares
func RunCalamaresInstaller(profile *pilot.Profile) {
	utils.LogCoala("%s[sysinstall]%s Preparazione motori...", utils.ColorCyan, utils.ColorReset)

	// 1. Fase Preparazione (Scrive gli script in /tmp/coa)
	// Questa fase non tocca /etc, quindi è sicura al 100%
	if err := calamares.PrepareOABootloader(profile); err != nil {
		utils.LogError("Errore preparazione script: %v", err)
		return
	}

	// 2. Fase Esecuzione (Pulisce /etc, estrae asset e lancia)
	// Questa funzione fa il "lavoro sporco" che ricordavi
	if err := calamares.SetupAndLaunch(); err != nil {
		utils.LogError("Calamares ha riscontrato un problema: %v", err)
		return
	}
}
