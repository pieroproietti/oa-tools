package calamares

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"coa/pkg/assets"
)

// Costanti globali del pacchetto calamares
const (
	coaCalamaresDir = "/etc/calamares"
	modulesDir      = "/etc/calamares/modules"
	stagingDir      = "/tmp/coa"
	ColorCyan       = "\033[1;36m"
	ColorReset      = "\033[0m"
)

// logCalamares stampa i log con il prefisso coa
func logCalamares(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[coa-calamares]%s %s\n", ColorCyan, ColorReset, msg)
}

// SetupAndLaunch coordina la pulizia, l'estrazione e l'avvio
func SetupAndLaunch() error {
	logCalamares("Generazione ambiente Evolution Edition...")

	os.RemoveAll(coaCalamaresDir)

	if err := assets.ExtractCalamares(coaCalamaresDir); err != nil {
		return fmt.Errorf("errore estrazione asset: %v", err)
	}
	
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return fmt.Errorf("errore creazione directory moduli: %v", err)
	}

	if err := deployDynamicConfigs(); err != nil {
		return err
	}

	logFile, _ := os.Create("/var/log/calamares.log")
	defer logFile.Close()

	cmd := exec.Command("calamares", "-d", "-D", "8") 
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	logCalamares("Avvio Calamares GUI...")
	return cmd.Run()
}

// deployDynamicConfigs sposta i file da /tmp/coa a /etc/calamares/modules
func deployDynamicConfigs() error {
	logCalamares("Configurazione moduli dinamici...")

	// UNPACKFS
	unpackConf := fmt.Sprintf("---\nunpack:\n  - source: \"%s\"\n    sourcefs: \"squashfs\"\n    destination: \"\"\n", findSquashfsPath())
	os.WriteFile(modulesDir+"/unpackfs.conf", []byte(unpackConf), 0644)

	// TRASLOCO MODULO OA-BOOTLOADER
	bootData, err := os.ReadFile(stagingDir + "/shellprocess_oa_bootloader.conf")
	if err != nil {
		return fmt.Errorf("modulo bootloader non trovato in staging: %v", err)
	}
	
	return os.WriteFile(modulesDir+"/shellprocess_oa_bootloader.conf", bootData, 0644)
}

// findSquashfsPath cerca il filesystem compresso
func findSquashfsPath() string {
	possiblePaths := []string{
		"/run/live/medium/live/filesystem.squashfs",
		"/lib/live/mount/medium/live/filesystem.squashfs",
		"/run/archiso/bootmnt/arch/x86_64/airootfs.sfs",
		"/run/initramfs/live/LiveOS/squashfs.img",
		"/live/filesystem.squashfs",
	}
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "/ERRORE_SQUASHFS_NON_TROVATO/filesystem.squashfs"
}
