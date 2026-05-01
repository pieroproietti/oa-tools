package calamares

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PrepareUserConf() error {
	// 1. Wishlist dei gruppi (standard Calamares)
	// Usiamo il formato semplice (stringa) o quello esteso (oggetto)
	wishlist := []string{"users", "wheel", "audio", "video", "storage", "network", "lp", "scanner"}

	// 2. Verifica dei gruppi esistenti nel sistema live
	data, err := os.ReadFile("/etc/group")
	if err != nil {
		return err
	}
	content := string(data)

	var yamlGroups string
	for _, g := range wishlist {
		if strings.Contains(content, g+":") {
			// Se il gruppo esiste, lo aggiungiamo
			yamlGroups += fmt.Sprintf("    - %s\n", g)
		}
	}

	// 3. Generiamo lo YAML basato sul tuo template
	config := fmt.Sprintf(`---
# OA-Tools: Configurazione Universale Dinamica
# Password: Approccio "Libertario" totale per Eggs & Bananas

defaultGroups:
%s

sudoersGroup:    wheel
sudoersConfigureWithGroup: false

# Disabilitiamo la rimozione dell'utente live qui se usiamo il modulo specifico
# removeLiveUser: true 

# --- IL FIX PER LE PASSWORD ---
# Permettiamo esplicitamente password deboli e lo impostiamo come default
allowWeakPasswords: true
allowWeakPasswordsDefault: true

passwordRequirements:
    minLength: -1
    maxLength: -1
    libpwquality:
        - minlen=0
        - minclass=0
        - dictcheck=0  # <--- DISATTIVA IL CONTROLLO DIZIONARIO
        - usercheck=0  # Disattiva controllo basato sul nome utente

# Configurazione User & Hostname
user:
  shell: /bin/bash
  forbidden_names: [ root, nobody ]
  home_permissions: "o700"

hostname:
  location: EtcFile
  writeHostsFile: true
  template: "oa-${product}"
`, yamlGroups)

	targetPath := "/etc/calamares/modules/users.conf"
	os.MkdirAll(filepath.Dir(targetPath), 0755)

	return os.WriteFile(targetPath, []byte(config), 0644)
}
