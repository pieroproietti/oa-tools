package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

// --- CONFIGURAZIONE ESPORTAZIONE ---
const (
	remoteUserHost = "root@192.168.1.2"
	remoteIsoPath  = "/var/lib/vz/template/iso/"
	remotePkgPath  = "/eggs/"
	isoSrcDir      = "/home/eggs"
)

// --- SISTEMA DI LOGGING INTERNO ---
// (Se hai già queste costanti in root.go o remaster.go, Go le riutilizzerà.
// Le ometto qui se sono già a livello di package, altrimenti assicurati di averle).
func logProcess(msg string) {
	fmt.Printf("\n\033[1;35m[PROCESS]\033[0m %s\n", msg)
}
func logClean(msg string) {
	fmt.Printf("\033[1;34m[CLEAN]\033[0m %s\n", msg)
}
func logCopy(msg string) {
	fmt.Printf("\033[1;34m[COPY]\033[0m %s\n", msg)
}
func logSuccess(msg string) {
	fmt.Printf("\033[1;32m[SUCCESS]\033[0m %s\n", msg)
}
func logError(msg string) {
	fmt.Printf("\033[1;31m[ERROR]\033[0m %s\n", msg)
}
func logWarning(msg string) {
	fmt.Printf("\033[1;33m[WARNING]\033[0m %s\n", msg)
}

// -----------------------------------

var cleanExport bool

var exportCmd = &cobra.Command{
	Use:    "export",
	Short:  "Export artifacts (iso, pkg) to a remote Proxmox storage",
	Long:   "Export generated ISOs or native packages to a remote server via SCP.",
	Hidden: false,
}

var exportIsoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Export the latest ISO to a remote Proxmox storage",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements("export iso", false)
		handleExportIso(cleanExport)
	},
}

var exportPkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Export the latest generated native package (.deb, .rpm, .pkg.tar.zst) to Proxmox",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements("export pkg", false)
		handleExportPkg(cleanExport)
	},
}

func init() {
	exportCmd.PersistentFlags().BoolVar(&cleanExport, "clean", false, "Clean old versions on remote server before exporting")

	exportCmd.AddCommand(exportIsoCmd)
	exportCmd.AddCommand(exportPkgCmd)
	rootCmd.AddCommand(exportCmd)
}

// =====================================================================
// LOGICA DI ESPORTAZIONE (Ex-Engine)
// =====================================================================

// handleExportIso copia in remoto la ISO utilizzando SSH/SCP con Multiplexing
func handleExportIso(clean bool) {
	allFiles, _ := filepath.Glob(filepath.Join(isoSrcDir, "egg-of_*.iso"))
	if len(allFiles) == 0 {
		logError(fmt.Sprintf("Nest is empty. No ISOs found in %s", isoSrcDir))
		return
	}

	latestFiles := make(map[string]string)
	re := regexp.MustCompile(`_\d{4}-\d{2}-\d{2}_\d{4}\.iso$`)

	for _, path := range allFiles {
		fileName := filepath.Base(path)
		prefix := re.ReplaceAllString(fileName, "")

		if info, err := os.Stat(path); err == nil {
			if current, exists := latestFiles[prefix]; exists {
				cInfo, _ := os.Stat(current)
				if info.ModTime().After(cInfo.ModTime()) {
					latestFiles[prefix] = path
				}
			} else {
				latestFiles[prefix] = path
			}
		}
	}

	// SSH Multiplexing setup
	socketPath := "/tmp/coa-ssh-mux"
	muxArgs := []string{
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=" + socketPath,
		"-o", "ControlPersist=2m",
	}

	defer func() {
		exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
		os.Remove(socketPath)
	}()

	for prefix, localPath := range latestFiles {
		targetFileName := filepath.Base(localPath)
		logProcess(fmt.Sprintf("Family: %s", prefix))

		if clean {
			logClean("Removing old versions on Proxmox...")
			rmCmdStr := fmt.Sprintf("rm -f %s%s*", remoteIsoPath, prefix)
			sshArgs := append(muxArgs, remoteUserHost, rmCmdStr)
			sshCmd := exec.Command("ssh", sshArgs...)
			sshCmd.Stdout, sshCmd.Stderr, sshCmd.Stdin = os.Stdout, os.Stderr, os.Stdin
			if err := sshCmd.Run(); err != nil {
				logWarning("Remote cleanup failed or no old files found.")
			} else {
				logSuccess("Old versions removed.")
			}
		}

		logCopy(fmt.Sprintf("Sending %s to Proxmox...", targetFileName))
		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remoteIsoPath)
		scpArgs := append(muxArgs, localPath, dstStr)
		scpCmd := exec.Command("scp", scpArgs...)
		scpCmd.Stdout, scpCmd.Stderr, scpCmd.Stdin = os.Stdout, os.Stderr, os.Stdin

		if err := scpCmd.Run(); err != nil {
			logError(fmt.Sprintf("Copy failed: %v", err))
		} else {
			logSuccess(fmt.Sprintf("%s is now on Proxmox.", targetFileName))
		}
	}
}

// handleExportPkg esporta i pacchetti nativi (DEB, Arch o RPM)
func handleExportPkg(clean bool) {
	logProcess("Searching for native packages...")

	// 1. Cerchiamo i pacchetti generati
	debFiles, _ := filepath.Glob("oa-tools*.deb")
	archFiles, _ := filepath.Glob("oa-tools*.pkg.tar.zst")
	rpmFiles, _ := filepath.Glob("oa-tools*.rpm")

	var allPackages []string
	allPackages = append(allPackages, debFiles...)
	allPackages = append(allPackages, archFiles...)
	allPackages = append(allPackages, rpmFiles...)

	if len(allPackages) == 0 {
		logError("No native packages found.")
		return
	}

	// --- SETUP SSH MULTIPLEXING ---
	socketPath := "/tmp/coa-ssh-mux-pkg"
	muxArgs := []string{
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=" + socketPath,
		"-o", "ControlPersist=2m",
	}

	defer func() {
		exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
		os.Remove(socketPath)
	}()

	if clean {
		logClean(fmt.Sprintf("Removing old oa-tools packages on %s...", remoteUserHost))

		// 2. Costruzione dinamica del target di pulizia
		var rmTargets string
		if len(debFiles) > 0 {
			rmTargets += remotePkgPath + "oa-tools*.deb "
		}
		if len(archFiles) > 0 {
			rmTargets += remotePkgPath + "oa-tools*.pkg.tar.zst "
		}
		if len(rpmFiles) > 0 {
			rmTargets += remotePkgPath + "oa-tools*.rpm "
		}

		cleanCmdStr := "rm -f " + rmTargets
		sshArgs := append(muxArgs, remoteUserHost, cleanCmdStr)

		cleanCmd := exec.Command("ssh", sshArgs...)
		cleanCmd.Stdout, cleanCmd.Stderr, cleanCmd.Stdin = os.Stdout, os.Stderr, os.Stdin

		if err := cleanCmd.Run(); err != nil {
			logError(fmt.Sprintf("Remote cleanup failed: %v", err))
		} else {
			logSuccess("Old packages removed.")
		}
	}

	// 3. Invio di tutti i pacchetti trovati
	for i, pkg := range allPackages {
		logCopy(fmt.Sprintf("Sending %s to Proxmox...", pkg))

		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remotePkgPath)
		scpArgs := append(muxArgs, pkg, dstStr)

		scpCmd := exec.Command("scp", scpArgs...)
		if i == 0 {
			scpCmd.Stdout, scpCmd.Stderr, scpCmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		} else {
			scpCmd.Stdout, scpCmd.Stderr = os.Stdout, os.Stderr
		}

		if err := scpCmd.Run(); err != nil {
			logError(fmt.Sprintf("SCP transfer failed for %s: %v", pkg, err))
		} else {
			logSuccess(fmt.Sprintf("%s successfully exported to Proxmox.", pkg))
		}
	}
}
