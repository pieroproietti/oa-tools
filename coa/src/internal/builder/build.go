package builder

import (
	"coa/src/internal/distro"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var AppVersion string

// parseGitVersion separa "0.6.2-5-g2504384" in (0.6.2, 5)
func parseGitVersion(v string) (string, string) {
	parts := strings.Split(v, "-")
	baseVer := parts[0]
	relNum := "1"

	if len(parts) > 1 {
		relNum = parts[1]
	}
	return baseVer, relNum
}

func generateDocs(coaDir string) error {
	// Definiamo i percorsi di output dentro il progetto
	docPath := filepath.Join(coaDir, "docs")
	manPath := filepath.Join(docPath, "man")
	compPath := filepath.Join(docPath, "completion")

	// Creiamo le directory se non esistono
	os.MkdirAll(manPath, 0755)
	os.MkdirAll(compPath, 0755)

	// Eseguiamo il binario coa appena compilato per generare i documenti
	// Nota: Assicurati che coa abbia un comando 'coa _gen_docs' o simile
	genCmd := exec.Command("./coa", "man", "--target", manPath)
	genCmd.Dir = coaDir
	if err := genCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate man pages: %w", err)
	}

	// Generazione completamenti (Bash, Zsh, Fish)
	shells := []string{"bash", "zsh", "fish"}
	for _, shell := range shells {
		targetFile := filepath.Join(compPath, "coa."+shell)
		compCmd := exec.Command("./coa", "completion", shell)
		compCmd.Dir = coaDir
		output, err := compCmd.Output()
		if err == nil {
			os.WriteFile(targetFile, output, 0644)
		}
	}

	return nil
}

func HandleBuild(d *distro.Distro, version string) {
	AppVersion = version
	baseVer, relNum := parseGitVersion(version)
	projRoot, oaDir, coaDir := getProjectPaths()

	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Println("\033[1;36m       COA BUILDER - Native Package Generation      \033[0m")
	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Printf("\033[1;34m[build]\033[0m Building version: %s\n", AppVersion)

	// 1. Compilazione motore C
	makeCmd := exec.Command("make", "-C", oaDir, fmt.Sprintf("VERSION=%s", AppVersion), "clean", "all")
	makeCmd.Stdout, makeCmd.Stderr = os.Stdout, os.Stderr
	if err := makeCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Engine compilation failed: %v\n", err)
		return
	}

	// 2. Compilazione orchestratore Go
	ldflags := fmt.Sprintf("-X 'coa/src/cmd.AppVersion=%s'", AppVersion)
	goCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", "coa", "./src")
	goCmd.Dir = coaDir
	goCmd.Stdout, goCmd.Stderr = os.Stdout, os.Stderr
	if err := goCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Orchestrator compilation failed: %v\n", err)
		return
	}

	// 3. Generazione Documentazione (Man pages & Completions)
	fmt.Println("\033[1;34m[build]\033[0m Generating documentation and completions...")
	if err := generateDocs(coaDir); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Docs generation failed: %v\n", err)
		return
	}

	// 3. Generazione pacchetto
	if d.FamilyID == "archlinux" {
		buildArchPackage(projRoot, baseVer, relNum)
	} else {
		pkgVersion := fmt.Sprintf("%s-%s", baseVer, relNum)
		buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion)
	}
}

func buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion string) {
	pkgName := fmt.Sprintf("oa-tools_%s_amd64", pkgVersion)
	buildDir := filepath.Join("/tmp", pkgName)

	// Pulizia preventiva
	os.RemoveAll(buildDir)

	// Creazione directory
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

	// Copia binari
	copyFile(filepath.Join(oaDir, "oa"), filepath.Join(buildDir, "usr/local/bin/oa"))
	copyFile(filepath.Join(coaDir, "coa"), filepath.Join(buildDir, "usr/local/bin/coa"))
	os.Chmod(filepath.Join(buildDir, "usr/local/bin/oa"), 0755)
	os.Chmod(filepath.Join(buildDir, "usr/local/bin/coa"), 0755)

	// Documentazione
	manDir := filepath.Join(buildDir, "usr/share/man/man1")
	exec.Command("sh", "-c", fmt.Sprintf("cp %s/docs/man/* %s/ && gzip -9 %s/*", coaDir, manDir, manDir)).Run()

	// Completamenti
	copyFile(filepath.Join(coaDir, "docs/completion/coa.bash"), filepath.Join(buildDir, "usr/share/bash-completion/completions/coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.zsh"), filepath.Join(buildDir, "usr/share/zsh/vendor-completions/_coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.fish"), filepath.Join(buildDir, "usr/share/fish/vendor_completions.d/coa.fish"))

	// CORREZIONE: Usa pkgVersion nel file control
	controlContent := fmt.Sprintf(`Package: oa-tools
Version: %s
Architecture: amd64
Maintainer: Piero Proietti <piero.proietti@gmail.com>
Depends: squashfs-tools, xorriso, live-boot, live-boot-initramfs-tools, dosfstools, mtools
Conflicts: penguins-eggs
Description: coa is the mind and oa the arm
`, pkgVersion)

	os.WriteFile(filepath.Join(buildDir, "DEBIAN", "control"), []byte(controlContent), 0644)

	fmt.Println("\033[1;34m[build]\033[0m Packing .deb archive...")
	dpkgCmd := exec.Command("dpkg-deb", "--build", buildDir)
	dpkgCmd.Stdout, dpkgCmd.Stderr = os.Stdout, os.Stderr
	dpkgCmd.Run()

	debFile := pkgName + ".deb"
	finalTarget := filepath.Join(projRoot, debFile)

	data, _ := os.ReadFile(filepath.Join("/tmp", debFile))
	os.WriteFile(finalTarget, data, 0644)

	os.RemoveAll(buildDir)
	os.Remove(filepath.Join("/tmp", debFile))

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Package created: \033[1m%s\033[0m\n", finalTarget)
}

func buildArchPackage(projRoot, baseVer, relNum string) {
	// Pulizia di eventuali spazi non standard o caratteri invisibili
	pkgbuildContent := fmt.Sprintf(`# Maintainer: Piero Proietti <piero.proietti@gmail.com>
pkgname=oa-tools
pkgver=%s
pkgrel=%s
pkgdesc="oa-tools universal Linux remastering"
arch=('x86_64')
license=('GPL3')
depends=('archiso' 'xorriso' 'squashfs-tools')
conflicts=('penguins-eggs')

package() {
    install -Dm755 "${startdir}/oa/oa" "${pkgdir}/usr/local/bin/oa"
    install -Dm755 "${startdir}/coa/coa" "${pkgdir}/usr/local/bin/coa"
    install -Dm644 "${startdir}/coa/docs/man/"* -t "${pkgdir}/usr/share/man/man1/"
    install -Dm644 "${startdir}/coa/docs/completion/coa.bash" "${pkgdir}/usr/share/bash-completion/completions/coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.zsh" "${pkgdir}/usr/share/zsh/vendor-completions/_coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/coa.fish"
}
`, baseVer, relNum)

	err := os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	if err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Failed to write PKGBUILD: %v\n", err)
		return
	}
	fmt.Printf("\033[1;32m[SUCCESS]\033[0m PKGBUILD generato: v%s rel%s\n", baseVer, relNum)
}

func getProjectPaths() (string, string, string) {
	cwd, _ := os.Getwd()
	projRoot := cwd
	if filepath.Base(cwd) == "coa" {
		projRoot = filepath.Dir(cwd)
	}
	return projRoot, filepath.Join(projRoot, "oa"), filepath.Join(projRoot, "coa")
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
