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

// HandleBuild coordina la compilazione e delega il packaging ai file specifici
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

	// 3. Generazione Documentazione
	fmt.Println("\033[1;34m[build]\033[0m Generating documentation and completions...")
	if err := generateDocs(coaDir); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Docs generation failed: %v\n", err)
		return
	}

	// 4. Routing verso i file specifici
	switch d.FamilyID {
	case "archlinux":
		//buildArchPackage(projRoot, coaDir, baseVer, relNum)
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
