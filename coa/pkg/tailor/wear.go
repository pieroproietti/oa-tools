package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
)

func Wear(costumeName string, noAcc bool, noFirm bool) error {
	root, err := getWardrobeRoot()
	if err != nil {
		return err
	}

	// 1. Identifichiamo se è un costume o un accessorio
	// Proviamo prima in costumes, poi in accessories
	costumeDir := filepath.Join(root, "costumes", costumeName)
	if _, err := os.Stat(costumeDir); os.IsNotExist(err) {
		costumeDir = filepath.Join(root, "accessories", costumeName)
	}

	// 2. Trova e carica lo YAML compatibile
	yamlFile := findCompatibleYaml(costumeDir)
	suit, err := loadSuit(yamlFile)
	if err != nil {
		return fmt.Errorf("impossibile caricare il costume %s: %v", costumeName, err)
	}

	utils.LogCoala("Indossando il costume: %s", suit.Name)

	// 3. Gestione Repositories (Update/Upgrade)
	if suit.Sequence.Repositories.Update {
		utils.LogCoala("Aggiornamento repository...")
		utils.ExecQuiet("apt-get update")
	}

	// 4. Installazione Pacchetti
	if len(suit.Sequence.Packages) > 0 {
		utils.LogCoala("Verifica e installazione pacchetti...")
		available := getAvailablePackages()
		var toInstall []string

		for _, pkg := range suit.Sequence.Packages {
			if _, ok := available[pkg]; ok {
				toInstall = append(toInstall, pkg)
			} else {
				utils.LogCoala("⚠️  Pacchetto non trovato, salto: %s", pkg)
			}
		}

		if len(toInstall) > 0 {
			installWithRetries(toInstall, 3)
		}
	}

	// 5. 🚀 LA PARACULATA: Rsync della sysroot
	// Se esiste la cartella 'sysroot' nel costume, la riversiamo in '/'
	sysrootPath := filepath.Join(costumeDir, "sysroot")
	if _, err := os.Stat(sysrootPath); err == nil {
		utils.LogCoala("Applicazione personalizzazioni (sysroot)...")
		// -a: archive, -H: hard-links, -S: sparse, -X: xattrs
		// Usiamo sudo implicitamente perché coa deve girare come root per queste operazioni
		cmd := fmt.Sprintf("rsync -aHSX %s/ /", sysrootPath)
		if err := utils.Exec(cmd); err != nil {
			utils.LogError("Errore durante l'applicazione della sysroot: %v", err)
		}
	}

	// 6. Finalizzazione (comandi personalizzati)
	if len(suit.Sequence.Cmds) > 0 {
		utils.LogCoala("Esecuzione comandi di sequenza...")
		for _, command := range suit.Sequence.Cmds {
			utils.Exec(command)
		}
	}

	if suit.Finalize.Customize && len(suit.Finalize.Cmds) > 0 {
		utils.LogCoala("Finalizzazione costume...")
		for _, command := range suit.Finalize.Cmds {
			utils.Exec(command)
		}
	}

	utils.LogCoala("✅ Costume '%s' indossato con successo!", suit.Name)

	if suit.Reboot {
		utils.LogCoala("🔄 Questo costume richiede un riavvio.")
	}

	return nil
}
