package calamares

import (
	"fmt"
	"os"
	"path/filepath"
)

func PrepareRemoveuserConf() error {
	// 1. Identifichiamo l'utente da rimuovere
	// Usiamo "live", che è il nostro standard per l'uovo
	liveUser := "live"

	// 2. Generiamo lo YAML specifico per il modulo removeuser
	// Nota: Il modulo si aspetta semplicemente la chiave 'username'
	config := fmt.Sprintf(`---
# OA-Tools: Configurazione Rimozione Utente Live
# Questo modulo assicura che l'utente '%s' non resti nel sistema installato

username: %s
`, liveUser, liveUser)

	// 3. Definiamo il percorso.
	targetPath := "/etc/calamares/modules/removeuser.conf"

	err := os.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(targetPath, []byte(config), 0644)
}
