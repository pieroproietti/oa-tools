package calamares

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GetActiveDM rileva quale Display Manager è installato nel "pieno"
func GetActiveDM() (string, string) {
	// Mappa dei DM e dei relativi file di configurazione
	dms := map[string]string{
		"lightdm": "/etc/lightdm/lightdm.conf",
		"sddm":    "/etc/sddm.conf",
		"gdm3":    "/etc/gdm3/daemon.conf",
		"gdm":     "/etc/gdm/custom.conf",
		"lxdm":    "/etc/lxdm/lxdm.conf",
	}

	for name, conf := range dms {
		if _, err := exec.LookPath(name); err == nil {
			return name, conf
		}
	}
	return "none", ""
}

// PrepareDisplaymanagerConf genera la configurazione per il login
func PrepareDisplaymanagerConf() error {
	dmName, _ := GetActiveDM()

	// Se non troviamo DM, meglio non scrivere il file o scriverlo vuoto
	if dmName == "none" {
		return nil
	}

	config := fmt.Sprintf(`---
# displaymanager.conf - Generato dinamicamente da coa v0.7.1
# Rilevato: %s

displaymanagers:
  - %s

# Configura l'autologin se l'utente lo ha richiesto nell'interfaccia
executable: "%s"
showAll: true
`, dmName, dmName, dmName)

	targetPath := "/etc/calamares/modules/displaymanager.conf"
	os.MkdirAll(filepath.Dir(targetPath), 0755)
	return os.WriteFile(targetPath, []byte(config), 0644)
}
