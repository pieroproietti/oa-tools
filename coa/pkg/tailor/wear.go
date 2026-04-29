package tailor

import (
	"coa/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
)

// Wear è il punto di ingresso principale
func Wear(costumeName string, noAcc bool, noFirm bool) error {
	utils.LogCoala("Inizio procedura di vestizione per: %s", costumeName)

	root, err := getWardrobeRoot()
	if err != nil {
		utils.LogError("Errore root guardaroba: %v", err)
		return err
	}

	costumeDir := filepath.Join(root, "costumes", costumeName)
	if _, err := os.Stat(costumeDir); os.IsNotExist(err) {
		return fmt.Errorf("costume '%s' non trovato in %s", costumeName, costumeDir)
	}

	yamlFile := findYaml(costumeDir)
	suit, err := loadSuit(yamlFile)
	if err != nil {
		return err
	}

	// 1. Applicazione del costume principale
	utils.LogCoala("--- Applicazione Costume: %s ---", suit.Name)
	if err := applySuit(costumeDir, suit); err != nil {
		return err
	}

	// 2. Applicazione Accessori (se presenti e non disabilitati)
	if !noAcc && len(suit.Accessories) > 0 {
		utils.LogCoala("--- Elaborazione %d accessori ---", len(suit.Accessories))
		for _, accName := range suit.Accessories {
			accDir := filepath.Join(root, "accessories", accName)
			if accYaml := findYaml(accDir); accYaml != "" {
				if accSuit, err := loadSuit(accYaml); err == nil {
					utils.LogCoala("Accessorio: %s", accName)
					applySuit(accDir, accSuit)
				}
			}
		}
	}

	// 3. Chiusura: Sincronizzazione home
	utils.LogCoala("--- Finalizzazione ---")
	copySkelToUser()

	utils.LogCoala("✅ Vestizione completata con successo!")
	return nil
}

// applySuit esegue le tre fasi: Pacchetti, Overlay, Comandi
func applySuit(dir string, suit *Suit) error {
	// Fase A: Pacchetti
	if len(suit.Packages) > 0 {
		utils.LogCoala("[%s] Tentativo installazione pacchetti: %v", suit.Name, suit.Packages)
		installWithRetries(suit.Packages, 3)
	} else {
		utils.LogCoala("[%s] Nessun pacchetto da installare.", suit.Name)
	}

	// Fase B: Overlay Sysroot
	sysrootPath := filepath.Join(dir, "sysroot")
	if _, err := os.Stat(sysrootPath); os.IsNotExist(err) {
		sysrootPath = filepath.Join(dir, "dirs") // Compatibilità col vecchio formato
	}

	if _, err := os.Stat(sysrootPath); err == nil {
		utils.LogCoala("[%s] Trovata cartella overlay: %s", suit.Name, sysrootPath)
		utils.LogCoala("[%s] Esecuzione rsync verso la radice /...", suit.Name)
		
		// Usiamo sudo rsync -aAXv per garantire il successo della "cucitura"
		cmd := fmt.Sprintf("sudo rsync -aAXv %s/ /", sysrootPath)
		if err := utils.Exec(cmd); err != nil {
			utils.LogCoala("[%s] Errore durante l'overlay: %v", suit.Name, err)
		} else {
			utils.LogCoala("[%s] Overlay completato correttamente.", suit.Name)
		}
	} else {
		utils.LogCoala("[%s] Nessuna cartella sysroot/dirs trovata, salto overlay.", suit.Name)
	}

	// Fase C: Comandi Post-Installazione
	if len(suit.Cmds) > 0 {
		utils.LogCoala("[%s] Esecuzione %d comandi post-installazione...", suit.Name, len(suit.Cmds))
		for _, command := range suit.Cmds {
			utils.LogCoala("[%s] Eseguo: %s", suit.Name, command)
			utils.Exec(command)
		}
	}

	return nil
}

// copySkelToUser sincronizza /etc/skel con la home dell'utente
func copySkelToUser() {
	userHome, _ := os.UserHomeDir()
	
	// Recuperiamo la home reale se siamo sotto sudo
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		userHome = filepath.Join("/home", sudoUser)
	}
	
	utils.LogCoala("Sincronizzazione /etc/skel -> %s", userHome)
	cmd := fmt.Sprintf("sudo rsync -a /etc/skel/ %s/", userHome)
	utils.Exec(cmd)
}
