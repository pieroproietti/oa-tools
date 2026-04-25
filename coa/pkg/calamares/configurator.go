package calamares

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"coa/pkg/assets"
)

const (
	coaCalamaresDir = "/etc/calamares"
	modulesDir      = "/etc/calamares/modules"
)

// --- SISTEMA DI LOGGING CALAMARES ---
const (
	ColorCyan  = "\033[1;36m"
	ColorRed   = "\033[1;31m"
	ColorReset = "\033[0m"
)

func logCalamares(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[coa-calamares]%s %s\n", ColorCyan, ColorReset, msg)
}

// ------------------------------------

// SetupAndLaunch prepara l'ambiente e lancia l'installer grafico
func SetupAndLaunch() error {
	logCalamares("Generazione ambiente universale...")

	// 1. Tabula rasa: pulizia della run precedente
	os.RemoveAll(coaCalamaresDir)

	// 2. Estrazione magica degli asset base (settings.conf, branding, ecc.) dal binario Go
	err := assets.ExtractCalamares(coaCalamaresDir)
	if err != nil {
		return fmt.Errorf("errore estrazione asset Calamares: %v", err)
	}

	// Assicuriamoci che la cartella modules esista fisicamente
	os.MkdirAll(modulesDir, 0755)

	// 3. Generazione del piano per il motore C (oa)
	// engine.GenerateFinalizePlan()

	// 4. CREAZIONE FILE DINAMICI

	// A. unpackfs.conf (trova dinamicamente lo squashfs)
	unpackConf := fmt.Sprintf("unpack:\n  - source: \"%s\"\n    sourcefs: \"squashfs\"\n    destination: \"\"\n", findSquashfsPath())
	os.WriteFile(modulesDir+"/unpackfs.conf", []byte(unpackConf), 0644)

	/**
	 * SCRIPT oa.sh
	 */
	// 1. Assicuriamoci che la directory base esista
	os.MkdirAll("/tmp/coa", 0755)

	// 2. Creiamo un vero script Bash fisico, blindato e con log
	wrapperScript := `#!/bin/bash
# Pulizia preventiva
mkdir -p /tmp/coa
rm -rf /tmp/coa/calamares-root

# Troviamo la vera cartella di mount di Calamares
REAL_ROOT=$(ls -d /tmp/calamares-root-* 2>/dev/null | head -n 1)

# Logghiamo cosa stiamo vedendo (VITALE per il debug!)
echo "=== AVVIO WRAPPER OA ===" > /var/log/oa-debug.log
echo "Root reale trovata: '$REAL_ROOT'" >> /var/log/oa-debug.log

if [ -z "$REAL_ROOT" ]; then
    echo "ERRORE CRITICO: Cartella calamares-root non trovata in /tmp!" >> /var/log/oa-debug.log
    exit 1
fi

# Creiamo il link simbolico assoluto e logghiamo il risultato
ln -sf "$REAL_ROOT" /tmp/coa/calamares-root
ls -la /tmp/coa/calamares-root >> /var/log/oa-debug.log

# Lanciamo il motore C
echo "Lancio oa..." >> /var/log/oa-debug.log
oa /tmp/coa/finalize-plan.json
`

	// Scriviamo lo script e lo rendiamo eseguibile
	os.WriteFile("/tmp/coa/run_oa.sh", []byte(wrapperScript), 0755)

	// 3. Modulo Calamares ridotto all'osso: esegue solo lo script
	finalizeConf := `dontChroot: true
timeout: 3600
script:
  - "/tmp/coa/run_oa.sh"
`
	os.WriteFile(modulesDir+"/shellprocess_oa_finalize.conf", []byte(finalizeConf), 0644)

	// LOG
	logFile, err := os.Create("/var/log/calamares.log")
	if err != nil {
		return fmt.Errorf("impossibile creare il file di log: %v", err)
	}
	defer logFile.Close()

	// Lanciamo Calamares in modalità super-debug (-d)
	cmd := exec.Command("calamares", "-d", coaCalamaresDir, "-D", "8")

	// Redirigiamo tutto: terminale + file di log
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	logCalamares("Avvio Calamares GUI...")
	return cmd.Run()
}

// findSquashfsPath cerca il file compresso del sistema (squashfs)
// nei percorsi standard usati dalle Live USB (Debian, Arch, Fedora, penguins-eggs).
// (L'ho resa privata con la 'f' minuscola visto che viene usata solo qui dentro)
func findSquashfsPath() string {
	possiblePaths := []string{
		"/run/live/medium/live/filesystem.squashfs",       // Debian Live moderna (e penguins-eggs su Debian)
		"/lib/live/mount/medium/live/filesystem.squashfs", // Debian Live vecchia
		"/run/archiso/bootmnt/arch/x86_64/airootfs.sfs",   // Arch Linux ISO standard
		"/run/initramfs/live/LiveOS/squashfs.img",         // Fedora Live standard
		"/live/filesystem.squashfs",                       // Fallback generico
	}

	for _, p := range possiblePaths {
		// Se il file esiste, restituiamo questo percorso
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Fallback estremo: se non trova nulla, restituisce un percorso
	// palesemente finto così Calamares fallisce in modo chiaro
	// invece di esplodere silenziosamente.
	return "/ERRORE_SQUASHFS_NON_TROVATO/filesystem.squashfs"
}
