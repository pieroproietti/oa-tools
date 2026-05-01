package cmd

import (
	"coa/pkg/calamares"
	"coa/pkg/pilot"
	"coa/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

var calamaresSubCmd = &cobra.Command{
	Use:   "calamares",
	Short: "Lancia l'installatore grafico Calamares",
	Run: func(cmd *cobra.Command, args []string) {
		// Verifichiamo i permessi prima di tutto
		CheckSudoRequirements("sysinstall calamares", true)

		// Lanciamo la logica di coordinamento
		RunCalamaresInstaller()
	},
}

// RunCalamaresInstaller coordina la preparazione e il lancio di Calamares
func RunCalamaresInstaller() {
	utils.LogCoala("%s[sysinstall]%s Preparazione motori...", utils.ColorCyan, utils.ColorReset)

	// 1. Caricamento del profilo tramite il Pilot
	profile, err := pilot.DetectAndLoad()
	if err != nil {
		utils.LogError("Impossibile caricare il profilo: %v", err)
		os.Exit(1)
	}

	// 2. Fase Preparazione (Scrive gli script in /tmp/coa)
	if err := calamares.PrepareOABootloader(profile); err != nil {
		utils.LogError("Errore preparazione script: %v", err)
		return
	}

	// 3. Fase di setuto (Pulisce /etc, estrae asse)
	// NOTA: Assicurati che SetupAndLaunch non pialli il file appena creato!
	if err := calamares.Setup(); err != nil {
		utils.LogError("Calamares ha riscontrato un problema: %v", err)
		return
	}

	// 4. Configurazione DINAMICA Utenti e Password
	// Lo facciamo qui per garantire che il file /etc/calamares/modules/users.conf
	// sia generato fresco in base alla distro corrente (Arch/Debian/Fedora)
	if err := calamares.PrepareUserConf(); err != nil {
		utils.LogError("Errore configurazione utenti: %v", err)
		// Non blocchiamo tutto, proviamo a procedere comunque
	}
	if err := calamares.PrepareRemoveuserConf(); err != nil {
		utils.LogError("Errore creazione removeuser.conf: %v", err)
		// Non blocchiamo tutto, proviamo a procedere comunque
	}

	// 5. LAUNCH: Calamares parte e trova la pappa pronta
	if err := calamares.Launch(); err != nil {
		utils.LogError("L'installatore si è chiuso con un errore: %v", err)
	}

}

func init() {
	// Appendiamo questo comando a sysinstallCmd
	sysinstallCmd.AddCommand(calamaresSubCmd)
}
