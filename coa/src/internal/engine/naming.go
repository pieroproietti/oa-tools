package engine

import (
	"coa/src/internal/distro"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// getIsoName genera il nome standard per l'immagine ISO finale
// Segue il formato: egg-of_distro-version-hostname_arch_timestamp.iso
func getIsoName(d *distro.Distro) string {
	hostname, _ := os.Hostname()
	timestamp := time.Now().Format("2006-01-02_1504")
	arch := runtime.GOARCH //

	var nameParts []string

	// 1. Identificativo della distribuzione (es. Debian, Arch)
	nameParts = append(nameParts, d.DistroID)

	// 2. Versione: priorità al Codename (es. bookworm) o alla Release (es. 41)
	if d.CodenameID != "" {
		nameParts = append(nameParts, d.CodenameID)
	} else if d.ReleaseID != "" {
		nameParts = append(nameParts, d.ReleaseID)
	}

	// 3. Nome dell'host per identificare la macchina di origine
	if hostname != "" {
		nameParts = append(nameParts, hostname)
	}

	// Uniamo le parti per creare il tag identificativo
	distroTag := strings.Join(nameParts, "-")

	// Restituiamo il nome completo del file ISO
	return fmt.Sprintf("egg-of_%s_%s_%s.iso", distroTag, arch, timestamp)
}
