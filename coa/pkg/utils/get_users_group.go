package utils

import (
	"os"
	"os/user"
)

// GetUserGroups recupera i gruppi dell'utente che ha invocato sudo
func GetUserGroups() []string {
	// 1. Identifichiamo l'utente originale tramite la variabile SUDO_USER
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		// Se non siamo sotto sudo, restituiamo dei default sicuri
		return []string{"wheel", "audio", "video", "storage", "network"}
	}

	u, err := user.Lookup(sudoUser)
	if err != nil {
		return []string{"wheel", "audio", "video", "storage"}
	}

	// 2. Recuperiamo tutti i GID dell'utente
	gids, err := u.GroupIds()
	if err != nil {
		return []string{"wheel", "audio", "video"}
	}

	var groups []string
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err == nil {
			// Escludiamo il gruppo primario (spesso uguale allo username)
			// per evitare conflitti durante la creazione dell'utente live
			if g.Name != u.Username && g.Name != "users" {
				groups = append(groups, g.Name)
			}
		}
	}

	// 3. Assicuriamoci che 'wheel' o 'sudo' siano presenti per i permessi
	groups = append(groups, "wheel")

	return groups
}
