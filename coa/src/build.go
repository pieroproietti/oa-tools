package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "embed" // L'underscore è fondamentale qui!
)

//go:embed VERSION
var rawVersion string

// Puliamo eventuali ritorni a capo (se premi Invio nel file di testo)
var AppVersion = strings.TrimSpace(rawVersion)

// getProjectPaths auto-rileva le directory corrette a prescindere da dove chiami 'coa'
func getProjectPaths() (string, string, string) {
	cwd, _ := os.Getwd()
	projRoot := cwd
	// Se "oa" non esiste nella cartella corrente, significa che siamo dentro "coa" o altrove. 
	// Saliamo di un livello per trovare la radice (oa-tools).
	if _, err := os.Stat(filepath.Join(cwd, "oa")); os.IsNotExist(err) {
		projRoot = filepath.Join(cwd, "..")
	}
	return projRoot, filepath.Join(projRoot, "oa"), filepath.Join(projRoot, "coa")
}

func handleBuild(d *Distro) {
	projRoot, oaDir, coaDir := getProjectPaths()

	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Println("\033[1;36m       COA BUILDER - Native Package Generation      \033[0m")
	fmt.Println("\033[1;36m====================================================\033[0m")

	// 1. Compilazione del motore C (oa)
	fmt.Println("\033[1;34m[build]\033[0m Compiling 'oa' C engine...")
	// makeCmd := exec.Command("make", "-C", oaDir, "clean", "all")
	makeCmd := exec.Command("make", "-C", oaDir, fmt.Sprintf("VERSION=%s", AppVersion), "clean", "all")
	makeCmd.Stdout = os.Stdout
	makeCmd.Stderr = os.Stderr
	if err := makeCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Engine compilation failed: %v\n", err)
		return
	}

	// 2. Compilazione dell'orchestratore Go (coa)
	fmt.Println("\033[1;34m[build]\033[0m Compiling 'coa' Go orchestrator...")
	goCmd := exec.Command("go", "build", "-o", "coa", "./src")
	goCmd.Dir = coaDir // <--- Forza l'esecuzione dentro la directory /coa
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	if err := goCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Orchestrator compilation failed: %v\n", err)
		return
	}

	// 3. Generazione Documentazione e Autocompletamento
	fmt.Println("\033[1;34m[build]\033[0m Generating man pages and completion scripts...")
	docsCmd := exec.Command("./coa", "docs")
	docsCmd.Dir = coaDir // <--- Forza l'esecuzione dentro la directory /coa
	docsCmd.Stdout = os.Stdout
	docsCmd.Stderr = os.Stderr
	if err := docsCmd.Run(); err != nil {
		fmt.Printf("\033[1;33m[WARNING]\033[0m Docs generation failed: %v\n", err)
	}

	// 4. Diramazione
	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Binaries and docs ready. Generating package for \033[1m%s\033[0m family...\n", d.FamilyID)

	switch d.FamilyID {
	case "debian":
		buildDebianPackage(projRoot, oaDir, coaDir)
	case "archlinux":
		buildArchPackage(projRoot)
	default:
		fmt.Printf("\033[1;33m[WARNING]\033[0m Automatic packaging for family '%s' is not yet implemented.\n", d.FamilyID)
	}
}

func buildDebianPackage(projRoot, oaDir, coaDir string) {
	pkgName := fmt.Sprintf("oa-tools_%s_amd64", AppVersion)
	buildDir := filepath.Join("/tmp", pkgName)

	// Struttura cartelle
	os.MkdirAll(filepath.Join(buildDir, "DEBIAN"), 0755)
	os.MkdirAll(filepath.Join(buildDir, "usr", "local", "bin"), 0755)
	
	manDir := filepath.Join(buildDir, "usr", "share", "man", "man1")
	bashDir := filepath.Join(buildDir, "usr", "share", "bash-completion", "completions")
	zshDir := filepath.Join(buildDir, "usr", "share", "zsh", "vendor-completions")
	fishDir := filepath.Join(buildDir, "usr", "share", "fish", "vendor_completions.d")
	
	os.MkdirAll(manDir, 0755)
	os.MkdirAll(bashDir, 0755)
	os.MkdirAll(zshDir, 0755)
	os.MkdirAll(fishDir, 0755)

	// Copia binari sicura con percorsi assoluti
	copyFile(filepath.Join(oaDir, "oa"), filepath.Join(buildDir, "usr", "local", "bin", "oa"))
	copyFile(filepath.Join(coaDir, "coa"), filepath.Join(buildDir, "usr", "local", "bin", "coa"))
	os.Chmod(filepath.Join(buildDir, "usr", "local", "bin", "oa"), 0755)
	os.Chmod(filepath.Join(buildDir, "usr", "local", "bin", "coa"), 0755)

	// Copia e compressione Man Pages
	manCmd := exec.Command("sh", "-c", fmt.Sprintf("cp ./docs/man/* %s/ && gzip -9 %s/*", manDir, manDir))
	manCmd.Dir = coaDir
	manCmd.Run()

	// Copia Autocompletamenti
	copyFile(filepath.Join(coaDir, "docs/completion/coa.bash"), filepath.Join(bashDir, "coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.zsh"), filepath.Join(zshDir, "_coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.fish"), filepath.Join(fishDir, "coa.fish"))

	// Control file
	controlContent := fmt.Sprintf(`Package: oa-tools
Version: %s
Architecture: amd64
Maintainer: Piero Proietti <piero.proietti@gmail.com>
Depends: squashfs-tools, xorriso
Description: The Artisan Orchestrator and Engine for Linux remastering.
 Mind and Body philosophy: coa (Go) and oa (C).
`, AppVersion)

	os.WriteFile(filepath.Join(buildDir, "DEBIAN", "control"), []byte(controlContent), 0644)

	fmt.Println("\033[1;34m[build]\033[0m Packing .deb archive...")
	dpkgCmd := exec.Command("dpkg-deb", "--build", buildDir)
	dpkgCmd.Stdout = os.Stdout
	dpkgCmd.Stderr = os.Stderr
	dpkgCmd.Run()

	// Spostamento del file finale nella root del progetto
	debFile := pkgName + ".deb"
	finalTarget := filepath.Join(projRoot, debFile)
	copyFile(filepath.Join("/tmp", debFile), finalTarget)
	os.RemoveAll(buildDir)
	os.Remove(filepath.Join("/tmp", debFile))

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Package created: \033[1m%s\033[0m\n", finalTarget)
}

func buildArchPackage(projRoot string) {
	pkgbuildContent := fmt.Sprintf(`# Maintainer: Piero Proietti <piero.proietti@gmail.com>
pkgname=oa-tools
pkgver=%s
pkgrel=1
pkgdesc="The Artisan Orchestrator and Engine for Linux remastering"
arch=('x86_64')
license=('GPL3')
depends=('xorriso' 'squashfs-tools')

package() {
    # Binaries
    install -Dm755 "${srcdir}/oa/oa" "${pkgdir}/usr/local/bin/oa"
    install -Dm755 "${srcdir}/coa/coa" "${pkgdir}/usr/local/bin/coa"

    # Man Pages
    install -Dm644 "${srcdir}/coa/docs/man/"* -t "${pkgdir}/usr/share/man/man1/"

    # Autocompletions
    install -Dm644 "${srcdir}/coa/docs/completion/coa.bash" "${pkgdir}/usr/share/bash-completion/completions/coa"
    install -Dm644 "${srcdir}/coa/docs/completion/coa.zsh" "${pkgdir}/usr/share/zsh/vendor-completions/_coa"
    install -Dm644 "${srcdir}/coa/docs/completion/coa.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/coa.fish"
}
`, AppVersion)

	os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	fmt.Printf("\033[1;32m[SUCCESS]\033[0m \033[1mPKGBUILD\033[0m generated in the project root.\n")
	fmt.Println("To package and install on Arch, run: \033[1mmakepkg -si\033[0m")
}
