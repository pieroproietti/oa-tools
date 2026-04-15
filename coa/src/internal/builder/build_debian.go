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

	// Pulizia preventiva
	os.RemoveAll(buildDir)

	// Creazione struttura directory standard Debian [cite: 41]
	dirs := []string{
		filepath.Join(buildDir, "DEBIAN"),
		filepath.Join(buildDir, "usr/local/bin"),
		filepath.Join(buildDir, "usr/share/man/man1"),
		filepath.Join(buildDir, "usr/share/bash-completion/completions"),
		filepath.Join(buildDir, "usr/share/zsh/vendor-completions"),
		filepath.Join(buildDir, "usr/share/fish/vendor_completions.d"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// 1. Installazione binari e creazione alias eggs
	copyFile(filepath.Join(oaDir, "oa"), filepath.Join(buildDir, "usr/local/bin/oa"))
	copyFile(filepath.Join(coaDir, "coa"), filepath.Join(buildDir, "usr/local/bin/coa"))
	os.Chmod(filepath.Join(buildDir, "usr/local/bin/oa"), 0755)
	os.Chmod(filepath.Join(buildDir, "usr/local/bin/coa"), 0755)

	// Link simbolico per il comando eggs
	os.Symlink("coa", filepath.Join(buildDir, "usr/local/bin/eggs"))

	// 2. Documentazione (Man pages) [cite: 41]
	manDir := filepath.Join(buildDir, "usr/share/man/man1")
	exec.Command("sh", "-c", fmt.Sprintf("cp %s/docs/man/*.1 %s/ && gzip -9 %s/*.1", coaDir, manDir, manDir)).Run()

	// 3. Completamenti shell e relativi alias
	bashTarget := filepath.Join(buildDir, "usr/share/bash-completion/completions/coa")
	copyFile(filepath.Join(coaDir, "docs/completion/coa.bash"), bashTarget)
	copyFile(filepath.Join(coaDir, "docs/completion/coa.zsh"), filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.fish"), filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/coa.fish"))

	// Symlink per i completamenti dell'alias eggs
	os.Symlink("coa", filepath.Join(buildDir, "usr/share/bash-completion/completions/eggs"))
	os.Symlink("_coa", filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_eggs"))
	os.Symlink("coa.fish", filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/eggs.fish"))

	// 4. FIX: Patch per l'autocompletamento Bash
	// Aggiungiamo la riga che associa la funzione di coa all'alias eggs
	f, err := os.OpenFile(bashTarget, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString("\n# eggs alias completion support\ncomplete -o default -F __start_coa eggs\n")
		f.Close()
	}

	// 5. Generazione file control [cite: 41]
	controlContent := fmt.Sprintf(`Package: oa-tools
Version: %s
Architecture: amd64
Maintainer: Piero Proietti <piero.proietti@gmail.com>
Depends: squashfs-tools, xorriso, live-boot, live-boot-initramfs-tools, dosfstools, mtools
Conflicts: penguins-eggs
Description: coa is the mind and oa the arm
`, pkgVersion)

	os.WriteFile(filepath.Join(buildDir, "DEBIAN", "control"), []byte(controlContent), 0644)

	// 6. Impacchettamento [cite: 41]
	fmt.Println("\033[1;34m[build]\033[0m Packing .deb archive...")
	dpkgCmd := exec.Command("dpkg-deb", "--build", buildDir)
	dpkgCmd.Stdout, dpkgCmd.Stderr = os.Stdout, os.Stderr
	dpkgCmd.Run()

	// Spostamento del pacchetto finale nella root del progetto
	debFile := pkgName + ".deb"
	finalTarget := filepath.Join(projRoot, debFile)
	data, _ := os.ReadFile(filepath.Join("/tmp", debFile))
	os.WriteFile(finalTarget, data, 0644)

	// Pulizia file temporanei
	os.RemoveAll(buildDir)
	os.Remove(filepath.Join("/tmp", debFile))

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Package created: \033[1m%s\033[0m\n", finalTarget)
}
