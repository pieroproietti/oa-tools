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

	// 3. LA CILIEGINA SULLA TORTA:
    // Ora che /etc/skel è completa di tutto (sfondi di colibri, config di eggs-dev, ecc.)
    // la copiamo nella home dell'utente corrente.
    if err := copySkelToUser(); err != nil {
        utils.LogError("Impossibile aggiornare la home utente: %v", err)
    }

	utils.LogCoala("✅ Sistema 'confezionato' con successo!")
	return nil
}

// applySuit esegue la sequenza operativa definita nello YAML
func applySuit(dir string, suit *Suit) error {
	utils.LogCoala("🧵 Cucitura componente: %s", suit.Name)

	// Directory per i path relativi (../../scripts/)
	originalWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(originalWd)

	// A. Update Repository
	if suit.Sequence.Repositories.Update {
		utils.LogCoala("[%s] Aggiornamento repository...", suit.Name)
		utils.ExecQuiet("apt-get update")
	}

	// B. Unione dei pacchetti (Root + Sequence)
	allPackages := append(suit.Packages, suit.Sequence.Packages...)
	if len(allPackages) > 0 {
		available := getAvailablePackages()
		var toInstall []string
		for _, pkg := range allPackages {
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

	// C. Sysroot / Dirs (La Paraculata) [cite: 2, 3, 9]
	sysrootPath := filepath.Join(dir, "sysroot")
	if _, err := os.Stat(sysrootPath); os.IsNotExist(err) {
		sysrootPath = filepath.Join(dir, "dirs")
	}

	if _, err := os.Stat(sysrootPath); err == nil {
		utils.LogCoala("🚀 Riversamento configurazioni per %s...", suit.Name)
		cmd := fmt.Sprintf("rsync -aHSX %s/ /", sysrootPath) // Il '/' finale è fondamentale 
		utils.Exec(cmd)
	}

	// D. Unione dei comandi (Root + Sequence + Finalize)
	allCmds := append(suit.Cmds, suit.Sequence.Cmds...)
	allCmds = append(allCmds, suit.Finalize.Cmds...)
	
	if len(allCmds) > 0 {
		utils.LogCoala("[%s] Esecuzione script...", suit.Name)
		for _, command := range allCmds {
			if strings.Contains(command, "hostname.sh") {
				// Ci assicuriamo di non duplicarlo se è già presente
				if !strings.Contains(command, suit.Name) {
					command = fmt.Sprintf("%s %s", command, suit.Name)
				}
			}
			utils.Exec(command)
		}
	}

	return nil
}

func copySkelToUser() error {
	// 1. Identifichiamo l'utente che ha lanciato sudo
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" || sudoUser == "root" {
		utils.LogCoala("Esecuzione non via sudo o come root, salto la copia della skel nella home.")
		return nil
	}

	// Recuperiamo la home directory dell'utente
	userHome := fmt.Sprintf("/home/%s", sudoUser)
	if sudoUser == "root" {
		userHome = "/root"
	}

	utils.LogCoala("Cucitura finale: applico le configurazioni desktop alla home di %s...", sudoUser)

	// 2. Usiamo rsync per copiare il contenuto di /etc/skel nella home dell'utente
	// -a: archive mode
	// Note: il '/' finale in /etc/skel/ è vitale per copiare il contenuto e non la cartella stessa
	rsyncCmd := fmt.Sprintf("rsync -a /etc/skel/ %s/", userHome)
	if err := utils.ExecQuiet(rsyncCmd); err != nil {
		return fmt.Errorf("errore durante la copia della skel: %v", err)
	}

	// 3. Sistemiamo i permessi (chown)
	// Essendo passati dalla sysroot del wardrobe a /etc/skel e poi alla home, 
	// dobbiamo assicurarci che l'utente possa leggere e scrivere i propri file.
	chownCmd := fmt.Sprintf("chown -R %s:%s %s", sudoUser, sudoUser, userHome)
	if err := utils.ExecQuiet(chownCmd); err != nil {
		return fmt.Errorf("errore durante il cambio di proprietà dei file: %v", err)
	}

	utils.LogCoala("✅ Configurazioni utente applicate correttamente.")
	return nil
}