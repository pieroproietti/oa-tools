package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Definiamo i colori per l'output in modo centralizzato.
// Questi blocchi rimarranno identici nei builder per Arch e Fedora.
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorBlue   = "\033[1;34m"
	ColorCyan   = "\033[36m"
	ColorYellow = "\033[33m"
)

func buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion string) {
	pkgName := fmt.Sprintf("oa-tools_%s_amd64", pkgVersion)
	buildDir := filepath.Join("/tmp", pkgName)

	// Pulizia preventiva del tavolo da lavoro
	os.RemoveAll(buildDir)

	// 1. Creazione struttura directory standard Debian.
	// Nota per Arch/Fedora: /usr/bin e /etc sono universali.
	// Le directory di man e completion potrebbero variare leggermente.
	dirs := []string{
		filepath.Join(buildDir, "DEBIAN"),
		filepath.Join(buildDir, "usr/bin"),
		filepath.Join(buildDir, "etc/oa-tools.d/brain.d"), // Configurazione modulare per coa (the mind)
		filepath.Join(buildDir, "usr/share/man/man1"),
		filepath.Join(buildDir, "usr/share/bash-completion/completions"),
		filepath.Join(buildDir, "usr/share/zsh/vendor-completions"),
		filepath.Join(buildDir, "usr/share/fish/vendor_completions.d"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// 2. Installazione binari e creazione alias 'eggs'.
	// Spostiamo tutto in /usr/bin per seguire gli standard delle distro moderne.
	binPath := filepath.Join(buildDir, "usr/bin")

	// Copiamo il Cervello (coa) e il Mulo (oa)
	copyFile(filepath.Join(oaDir, "oa"), filepath.Join(binPath, "oa"))
	copyFile(filepath.Join(coaDir, "coa"), filepath.Join(binPath, "coa"))

	os.Chmod(filepath.Join(binPath, "oa"), 0755)
	os.Chmod(filepath.Join(binPath, "coa"), 0755)

	// Link simbolico per compatibilità: 'eggs' ora punta al nuovo cuore 'coa'
	os.Symlink("coa", filepath.Join(binPath, "eggs"))

	// 3. Gestione della Configurazione YAML (/etc/oa-tools.d).
	// Questo blocco definisce l'identità del sistema e sarà il riferimento per ogni distro.
	confDest := filepath.Join(buildDir, "etc/oa-tools.d")
	
	// Generazione del file di configurazione principale.
	// Il dialetto è "oa" [cite: 30-03-2026] e la filosofia è integrata [cite: 29-03-2026].
	oaYamlContent := fmt.Sprintf(`---
# oa-tools configuration
# coa is the mind and oa the arm
# Philosophy: https://penguins-eggs.net/blog/eggs-bananas

system:
  dialect: "%s"
  version: "%s"

wardrobe:
  root: "~/.oa-wardrobe"
  repo: "https://github.com/pieroproietti/oa-wardrobe.git"

remaster:
  default_user: "artisan"
  work_dir: "/home/eggs"
`, "oa", pkgVersion)

	os.WriteFile(filepath.Join(confDest, "oa-tools.yaml"), []byte(oaYamlContent), 0644)

	// Riversamento di eventuali file dalla cartella 'conf' del progetto (es. brain.d/*.yaml)
	confSrc := filepath.Join(projRoot, "conf")
	if _, err := os.Stat(confSrc); err == nil {
		exec.Command("sh", "-c", fmt.Sprintf("cp -r %s/* %s/", confSrc, confDest)).Run()
	}

	// 4. Documentazione (Man pages)
	manDir := filepath.Join(buildDir, "usr/share/man/man1")
	exec.Command("sh", "-c", fmt.Sprintf("cp %s/docs/man/*.1 %s/ && gzip -9 %s/*.1", coaDir, manDir, manDir)).Run()

	// 5. Completamenti shell e relativi alias
	bashTarget := filepath.Join(buildDir, "usr/share/bash-completion/completions/coa")
	copyFile(filepath.Join(coaDir, "docs/completion/coa.bash"), bashTarget)
	copyFile(filepath.Join(coaDir, "docs/completion/coa.zsh"), filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.fish"), filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/coa.fish"))

	// Symlink per i completamenti dell'alias eggs per tutte le shell
	os.Symlink("coa", filepath.Join(buildDir, "usr/share/bash-completion/completions/eggs"))
	os.Symlink("_coa", filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_eggs"))
	os.Symlink("coa.fish", filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/eggs.fish"))

	// 6. Patch per l'autocompletamento Bash dell'alias
	f, err := os.OpenFile(bashTarget, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString("\n# eggs alias completion support\ncomplete -o default -F __start_coa eggs\n")
		f.Close()
	}

	// 7. Generazione file control (Debian Specific)
	// Nota per Arch: i dati qui sotto andranno nel PKGBUILD.
	// Nota per Fedora: i dati qui sotto andranno nello SPEC file.
	controlContent := fmt.Sprintf(`Package: oa-tools
Version: %s
Architecture: amd64
Maintainer: Piero Proietti <piero.proietti@gmail.com>
Depends: squashfs-tools, xorriso, live-boot, live-boot-initramfs-tools, dosfstools, mtools, rsync, git, sudo
Conflicts: penguins-eggs
Description: coa is the mind and oa the arm
`, pkgVersion)

	os.WriteFile(filepath.Join(buildDir, "DEBIAN", "control"), []byte(controlContent), 0644)

	// 8. Impacchettamento finale
	fmt.Printf("%s[build]%s Packing .deb archive (%s)...\n", ColorBlue, ColorReset, pkgVersion)
	dpkgCmd := exec.Command("dpkg-deb", "--build", buildDir)
	if err := dpkgCmd.Run(); err != nil {
		fmt.Printf("%s[ERROR]%s Failed to build package: %v\n", ColorReset, ColorReset, err)
		return
	}

	// Spostamento del pacchetto finale nella root del progetto
	debFile := pkgName + ".deb"
	finalTarget := filepath.Join(projRoot, debFile)

	data, _ := os.ReadFile(filepath.Join("/tmp", debFile))
	os.WriteFile(finalTarget, data, 0644)

	// Pulizia finale per lasciare il sistema in ordine
	os.RemoveAll(buildDir)
	os.Remove(filepath.Join("/tmp", debFile))

	fmt.Printf("%s[SUCCESS]%s Package created: %s\n", ColorGreen, ColorReset, finalTarget)
}
