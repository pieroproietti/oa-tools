package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion string) {
	pkgName := fmt.Sprintf("oa-tools_%s_amd64", pkgVersion)
	buildDir := filepath.Join("/tmp", pkgName)

	// Pulizia preventiva del tavolo da lavoro
	os.RemoveAll(buildDir)

	// 1. Creazione struttura directory standard Debian.
	dirs := []string{
		filepath.Join(buildDir, "DEBIAN"),
		filepath.Join(buildDir, "usr/bin"),
		filepath.Join(buildDir, "etc/oa-tools.d/brain.d"), // Cartella per i profili delle distribuzioni
		filepath.Join(buildDir, "usr/share/man/man1"),
		filepath.Join(buildDir, "usr/share/bash-completion/completions"),
		filepath.Join(buildDir, "usr/share/zsh/vendor-completions"),
		filepath.Join(buildDir, "usr/share/fish/vendor_completions.d"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// 2. Installazione binari e creazione alias 'eggs'.
	binPath := filepath.Join(buildDir, "usr/bin")
	copyFile(filepath.Join(oaDir, "oa"), filepath.Join(binPath, "oa"))
	copyFile(filepath.Join(coaDir, "coa"), filepath.Join(binPath, "coa"))
	os.Chmod(filepath.Join(binPath, "oa"), 0755)
	os.Chmod(filepath.Join(binPath, "coa"), 0755)
	os.Symlink("coa", filepath.Join(binPath, "eggs"))

	// 3. Gestione della Configurazione YAML e Popolamento brain.d
	confDest := filepath.Join(buildDir, "etc/oa-tools.d")
	brainDest := filepath.Join(confDest, "brain.d")
	
	// Generazione del file di configurazione principale oa-tools.yaml
	oaYamlContent := fmt.Sprintf(`---
# oa-tools configuration
# coa is the mind and oa the arm
# Philosophy: https://penguins-eggs.net/blog/eggs-bananas

system:
  dialect: "oa"
  version: "%s"

wardrobe:
  root: "~/.oa-wardrobe"
  repo: "https://github.com/pieroproietti/oa-wardrobe.git"

remaster:
  default_user: "artisan"
  work_dir: "/home/eggs"
`, pkgVersion)

	os.WriteFile(filepath.Join(confDest, "oa-tools.yaml"), []byte(oaYamlContent), 0644)

	// --- IL FIX: Copia del contenuto di brain.d ---
	// Prendiamo l'indice e i profili YAML dalla cartella sorgente di sviluppo
	brainSrc := filepath.Join(projRoot, "coa", "brain.d")
	if _, err := os.Stat(brainSrc); err == nil {
		fmt.Printf("%s[build]%s Popolamento brain.d con i profili distro...\n", ColorBlue, ColorReset)
		// Copiamo tutto il contenuto (*.yaml) nella destinazione del pacchetto
		exec.Command("sh", "-c", fmt.Sprintf("cp -r %s/* %s/", brainSrc, brainDest)).Run()
	} else {
		fmt.Printf("%s[warning]%s Sorgente brain.d non trovata in %s\n", ColorYellow, ColorReset, brainSrc)
	}

	// 4. Documentazione (Man pages)
	manDir := filepath.Join(buildDir, "usr/share/man/man1")
	exec.Command("sh", "-c", fmt.Sprintf("cp %s/docs/man/*.1 %s/ && gzip -9 %s/*.1", coaDir, manDir, manDir)).Run()

	// 5. Completamenti shell e relativi alias
	bashTarget := filepath.Join(buildDir, "usr/share/bash-completion/completions/coa")
	copyFile(filepath.Join(coaDir, "docs/completion/coa.bash"), bashTarget)
	copyFile(filepath.Join(coaDir, "docs/completion/coa.zsh"), filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.fish"), filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/coa.fish"))

	os.Symlink("coa", filepath.Join(buildDir, "usr/share/bash-completion/completions/eggs"))
	os.Symlink("_coa", filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_eggs"))
	os.Symlink("coa.fish", filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/eggs.fish"))

	// 6. Patch per l'autocompletamento Bash dell'alias
	f, err := os.OpenFile(bashTarget, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString("\n# eggs alias completion support\ncomplete -o default -F __start_coa eggs\n")
		f.Close()
	}

	// 7. Generazione file control
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
		fmt.Printf("%s[ERROR]%s Failed to build package: %v\n", ColorRed, ColorReset, err)
		return
	}

	// Spostamento del pacchetto finale nella root del progetto
	debFile := pkgName + ".deb"
	finalTarget := filepath.Join(projRoot, debFile)
	data, _ := os.ReadFile(filepath.Join("/tmp", debFile))
	os.WriteFile(finalTarget, data, 0644)

	// Pulizia finale
	os.RemoveAll(buildDir)
	os.Remove(filepath.Join("/tmp", debFile))

	fmt.Printf("%s[SUCCESS]%s Package created: %s\n", ColorGreen, ColorReset, finalTarget)
}
