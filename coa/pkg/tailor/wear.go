package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Wear(costumeName string, noAcc bool, noFirm bool) error {
	root, err := getWardrobeRoot()
	if err != nil {
		return err
	}

	// 1. Caricamento del Costume principale
	costumeDir := filepath.Join(root, "costumes", costumeName)
	if _, err := os.Stat(costumeDir); os.IsNotExist(err) {
		return fmt.Errorf("costume '%s' non trovato", costumeName)
	}

	yamlFile := findCompatibleYaml(costumeDir)
	mainSuit, err := loadSuit(yamlFile)
	if err != nil {
		return err
	}

	utils.LogCoala("Sarto all'opera: inizio con il costume principale %s", mainSuit.Name)

	// 2. 🧵 PRIMA IL COSTUME (La base del sistema)
	if err := applySuit(costumeDir, mainSuit); err != nil {
		return fmt.Errorf("errore critico durante l'installazione del costume: %v", err)
	}

	// 3. 🎩 DOPO GLI ACCESSORI (Strumenti aggiuntivi)
	if !noAcc && len(mainSuit.Accessories) > 0 {
		utils.LogCoala("Configurazione base completata. Passo agli accessori...")

		for _, accName := range mainSuit.Accessories {
			utils.LogCoala("📦 Installazione accessorio: %s", accName)

			accDir := filepath.Join(root, "accessories", accName)
			if _, err := os.Stat(accDir); os.IsNotExist(err) {
				utils.LogCoala("⚠️  Accessorio '%s' non trovato, salto.", accName)
				continue
			}

			accYaml := findCompatibleYaml(accDir)
			accSuit, err := loadSuit(accYaml)
			if err != nil {
				utils.LogCoala("⚠️  Errore caricamento accessorio %s: %v", accName, err)
				continue
			}

			// Applichiamo l'accessorio
			if err := applySuit(accDir, accSuit); err != nil {
				utils.LogCoala("❌ Errore durante l'applicazione di %s: %v", accName, err)
			}
		}
	}

	utils.LogCoala("✅ Sistema 'confezionato' con successo!")
	return nil
}

// applySuit esegue la sequenza operativa definita nello YAML
func applySuit(dir string, suit *Suit) error {
	utils.LogCoala("🧵 Cucitura componente: %s", suit.Name)

	// 🚀 Cambio Directory per risolvere i path relativi degli script (../../scripts/)
	originalWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(originalWd)

	// A. Repository Update
	if suit.Sequence.Repositories.Update {
		utils.LogCoala("[%s] Aggiornamento repository...", suit.Name)
		utils.ExecQuiet("apt-get update")
	}

	// B. Installazione Pacchetti
	if len(suit.Sequence.Packages) > 0 {
		available := getAvailablePackages() // Ora usa la logica stream corretta
		var toInstall []string
		for _, pkg := range suit.Sequence.Packages {
			cleanPkg := strings.TrimSpace(pkg)
			if _, ok := available[cleanPkg]; ok {
				toInstall = append(toInstall, cleanPkg)
			} else {
				utils.LogCoala("⚠️  [%s] Pacchetto non trovato: %s", suit.Name, cleanPkg)
			}
		}
		if len(toInstall) > 0 {
			installWithRetries(toInstall, 3)
		}
	}

	// C. La Paraculata (Sysroot)
	// Controlliamo se esiste una cartella 'sysroot' o 'dirs' nell'accessorio/costume
	sysrootPath := filepath.Join(dir, "sysroot")
	if _, err := os.Stat(sysrootPath); os.IsNotExist(err) {
		// Alcuni tuoi accessori usano 'dirs' invece di 'sysroot'
		sysrootPath = filepath.Join(dir, "dirs")
	}

	if _, err := os.Stat(sysrootPath); err == nil {
		utils.LogCoala("🚀 Riversamento sysroot/dirs per %s...", suit.Name)
		cmd := fmt.Sprintf("rsync -aHSX %s/ /", sysrootPath)
		utils.Exec(cmd)
	}

	// D. Esecuzione Script (Sequence e Finalize)
	allCmds := append(suit.Sequence.Cmds, suit.Finalize.Cmds...)
	if len(allCmds) > 0 {
		utils.LogCoala("[%s] Esecuzione script di configurazione...", suit.Name)
		for _, command := range allCmds {
			utils.Exec(command)
		}
	}

	return nil
}
