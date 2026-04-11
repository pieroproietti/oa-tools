package builder

import (
	"coa/src/internal/distro"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AppVersion deve essere accessibile (possiamo passarla o usare cmd.AppVersion)
var AppVersion string

// parseGitVersion separa "0.6.2-5-g2504384" in (0.6.2, 5)
func parseGitVersion(v string) (string, string) {
	parts := strings.Split(v, "-")

	baseVer := parts[0]
	relNum := "1" // Default se non ci sono commit extra

	if len(parts) > 1 {
		relNum = parts[1]
	}

	return baseVer, relNum
}

// HandleBuild orchestra la compilazione e la creazione del pacchetto nativo
func HandleBuild(d *distro.Distro, version string) {
	AppVersion = version
	baseVer, relNum := parseGitVersion(version)

	projRoot, oaDir, coaDir := getProjectPaths()

	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Println("\033[1;36m       COA BUILDER - Native Package Generation      \033[0m")
	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Printf("\033[1;34m[build]\033[0m Building version: %s\n", AppVersion)

	// 1. Compilazione del motore C (oa)
	fmt.Println("\033[1;34m[build]\033[0m Compiling 'oa' C engine...")
	makeCmd := exec.Command("make", "-C", oaDir, fmt.Sprintf("VERSION=%s", AppVersion), "clean", "all")
	makeCmd.Stdout = os.Stdout
	makeCmd.Stderr = os.Stderr
	if err := makeCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Engine compilation failed: %v\n", err)
		return
	}

	// 2. Compilazione dell'orchestratore Go (coa)
	fmt.Println("\033[1;34m[build]\033[0m Compiling 'coa' Go orchestrator...")
	ldflags := fmt.Sprintf("-X 'coa/src/cmd.AppVersion=%s'", AppVersion)
	goCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", "coa", "./src")
	goCmd.Dir = coaDir
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	if err := goCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Orchestrator compilation failed: %v\n", err)
		return
	}

	// 3. Generazione pacchetto in base alla famiglia
	if d.FamilyID == "archlinux" {
		buildArchPackage(projRoot, baseVer, relNum)
	} else {
		// Debian vuole il formato "versione-release"
		pkgVersion := fmt.Sprintf("%s-%s", baseVer, relNum)
		buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion)
	}
}
func buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion string) {
	// pkgVersion qui è già "0.6.2-5"
	pkgName := fmt.Sprintf("oa-tools_%s_amd64", pkgVersion)
	buildDir := filepath.Join("/tmp", pkgName)

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

	copyFile(filepath.Join(oaDir, "oa"), filepath.Join(buildDir, "usr", "local", "bin", "oa"))
	copyFile(filepath.Join(coaDir, "coa"), filepath.Join(buildDir, "usr", "local", "bin", "coa"))
	os.Chmod(filepath.Join(buildDir, "usr", "local", "bin", "oa"), 0755)
	os.Chmod(filepath.Join(buildDir, "usr", "local", "bin", "coa"), 0755)

	manCmd := exec.Command("sh", "-c", fmt.Sprintf("cp %s/docs/man/* %s/ && gzip -9 %s/*", coaDir, manDir, manDir))
	manCmd.Run()

	copyFile(filepath.Join(coaDir, "docs/completion/coa.bash"), filepath.Join(bashDir, "coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.zsh"), filepath.Join(zshDir, "_coa"))
	copyFile(filepath.Join(coaDir, "docs/completion/coa.fish"), filepath.Join(fishDir, "coa.fish"))

	controlContent := fmt.Sprintf(`Package: oa-tools
Version: %s
Architecture: amd64
Maintainer: Piero Proietti <piero.proietti@gmail.com>
Depends: squashfs-tools, xorriso, live-boot, live-boot-initramfs-tools, dosfstools, mtools
Conflicts: penguins-eggs
Description: coa is the mind and oa the arm
`, AppVersion)

	os.WriteFile(filepath.Join(buildDir, "DEBIAN", "control"), []byte(controlContent), 0644)

	fmt.Println("\033[1;34m[build]\033[0m Packing .deb archive...")
	dpkgCmd := exec.Command("dpkg-deb", "--build", buildDir)
	dpkgCmd.Stdout = os.Stdout
	dpkgCmd.Stderr = os.Stderr
	dpkgCmd.Run()

	debFile := pkgName + ".deb"
	finalTarget := filepath.Join(projRoot, debFile)

	// Usiamo os.Rename o una copia manuale se sono su filesystem diversi
	data, _ := os.ReadFile(filepath.Join("/tmp", debFile))
	os.WriteFile(finalTarget, data, 0644)

	os.RemoveAll(buildDir)
	os.Remove(filepath.Join("/tmp", debFile))

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Package created: \033[1m%s\033[0m\n", finalTarget)
}

func buildArchPackage(projRoot, baseVer, relNum string) {
	pkgbuildContent := fmt.Sprintf(`# Maintainer: Piero Proietti <piero.proietti@gmail.com> 
pkgname=oa-tools
pkgver=%s
pkgrel=%s
pkgdesc="oa-tools universal Linux remastering" 
arch=('x86_64')
license=('GPL3')
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
`, AppVersion)

	os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	fmt.Printf("\033[1;32m[SUCCESS]\033[0m \033[1mPKGBUILD\033[0m generated in the project root.\n")
}

func getProjectPaths() (string, string, string) {
	cwd, _ := os.Getwd()
	projRoot := cwd
	// Se siamo dentro coa/, risaliamo
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
