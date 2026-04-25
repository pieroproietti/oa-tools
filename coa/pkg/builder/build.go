package builder

import (
	"coa/pkg/distro"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// --- SISTEMA DI LOGGING DEL BUILDER ---
const (
	ColorCyan  = "\033[1;36m"
	ColorBlue  = "\033[1;34m"
	ColorRed   = "\033[1;31m"
	ColorReset = "\033[0m"
)

func LogBuild(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[build]%s %s\n", ColorBlue, ColorReset, msg)
}

func LogError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[ERROR]%s %s\n", ColorRed, ColorReset, msg)
}

// --------------------------------------

var AppVersion string

// HandleBuild coordina la compilazione e delega il packaging ai file specifici
func HandleBuild(d *distro.Distro, version string) {
	AppVersion = version
	baseVer, relNum := parseGitVersion(version)
	projRoot, oaDir, coaDir := getProjectPaths()

	// Header pulito con costanti
	fmt.Printf("%s====================================================%s\n", ColorCyan, ColorReset)
	fmt.Printf("%s       COA BUILDER - Native Package Generation      %s\n", ColorCyan, ColorReset)
	fmt.Printf("%s====================================================%s\n", ColorCyan, ColorReset)

	LogBuild("Building version: %s", AppVersion)

	// 1. Compilazione motore C
	makeCmd := exec.Command("make", "-C", oaDir, fmt.Sprintf("VERSION=%s", AppVersion), "clean", "all")
	makeCmd.Stdout, makeCmd.Stderr = os.Stdout, os.Stderr
	if err := makeCmd.Run(); err != nil {
		LogError("Engine compilation failed: %v", err)
		return
	}

	// 2. Compilazione orchestratore Go
	// FIX: Sostituito "./pkg" con "main.go" per allinearlo al nuovo Makefile
	ldflags := fmt.Sprintf("-X 'coa/pkg/cmd.AppVersion=%s'", AppVersion)
	goCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", "coa", "main.go")
	goCmd.Dir = coaDir
	goCmd.Stdout, goCmd.Stderr = os.Stdout, os.Stderr
	if err := goCmd.Run(); err != nil {
		LogError("Orchestrator compilation failed: %v", err)
		return
	}

	// 3. Generazione Documentazione
	LogBuild("Generating documentation and completions...")
	if err := generateDocs(coaDir); err != nil {
		LogError("Docs generation failed: %v", err)
		return
	}

	// 4. Routing verso i file specifici
	switch d.FamilyID {
	case "archlinux":
		buildArchPackage(projRoot, baseVer, relNum)
	case "fedora", "rhel", "centos", "rocky", "almalinux":
		buildFedoraPackage(projRoot, oaDir, coaDir, baseVer, relNum)
	default:
		pkgVersion := fmt.Sprintf("%s-%s", baseVer, relNum)
		buildDebianPackage(projRoot, oaDir, coaDir, pkgVersion)
	}
}

// Utility condivise

func parseGitVersion(v string) (string, string) {
	cleanV := strings.TrimPrefix(v, "v")
	parts := strings.Split(cleanV, "-")
	baseVer := parts[0]
	relNum := "1"
	if len(parts) > 1 {
		relNum = parts[1]
	}
	return baseVer, relNum
}

func generateDocs(coaDir string) error {
	docPath := filepath.Join(coaDir, "docs")
	genCmd := exec.Command("./coa", "_gen_docs", "--target", docPath)
	genCmd.Dir = coaDir
	genCmd.Stdout, genCmd.Stderr = os.Stdout, os.Stderr
	return genCmd.Run()
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
