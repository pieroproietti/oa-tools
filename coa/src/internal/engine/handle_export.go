package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

const (
	remoteUserHost = "root@192.168.1.2"
	remoteIsoPath  = "/var/lib/vz/template/iso/"
	remotePkgPath  = "/eggs/"
	isoSrcDir      = "/home/eggs"
)

// HandleExportIso copia in remoto la ISO utilizzando SSH/SCP con Multiplexing
func HandleExportIso(clean bool) {
	allFiles, _ := filepath.Glob(filepath.Join(isoSrcDir, "egg-of_*.iso"))
	if len(allFiles) == 0 {
		fmt.Println("\033[1;31m[coa]\033[0m Nest is empty. No ISOs found in", isoSrcDir)
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
		fmt.Printf("\n\033[1;35m[PROCESS]\033[0m Family: %s\n", prefix)

		if clean {
			fmt.Printf("\033[1;34m[CLEAN]\033[0m Removing old versions on Proxmox...\n")
			rmCmdStr := fmt.Sprintf("rm -f %s%s*", remoteIsoPath, prefix)
			sshArgs := append(muxArgs, remoteUserHost, rmCmdStr)
			sshCmd := exec.Command("ssh", sshArgs...)
			sshCmd.Stdout, sshCmd.Stderr, sshCmd.Stdin = os.Stdout, os.Stderr, os.Stdin
			if err := sshCmd.Run(); err != nil {
				fmt.Printf("\033[1;33m[WARNING]\033[0m Remote cleanup failed or no old files found.\n")
			} else {
				fmt.Printf("\033[1;32m[CLEAN]\033[0m Old versions removed.\n")
			}
		}

		fmt.Printf("\033[1;32m[COPY]\033[0m Sending %s to Proxmox...\n", targetFileName)
		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remoteIsoPath)
		scpArgs := append(muxArgs, localPath, dstStr)
		scpCmd := exec.Command("scp", scpArgs...)
		scpCmd.Stdout, scpCmd.Stderr, scpCmd.Stdin = os.Stdout, os.Stderr, os.Stdin

		if err := scpCmd.Run(); err != nil {
			fmt.Printf("\033[1;31m[ERROR]\033[0m Copy failed: %v\n", err)
		} else {
			fmt.Printf("\033[1;32m[SUCCESS]\033[0m %s is now on Proxmox.\n", targetFileName)
		}
	}
}

// HandleExportPkg esporta i pacchetti nativi (DEB, Arch o RPM)
func HandleExportPkg(clean bool) {
	fmt.Println("\033[1;34m[PROCESS]\033[0m Searching for native packages...")

	// 1. Cerchiamo i pacchetti generati (DEB, Arch e ora RPM)
	debFiles, _ := filepath.Glob("oa-tools*.deb")
	archFiles, _ := filepath.Glob("oa-tools*.pkg.tar.zst")
	rpmFiles, _ := filepath.Glob("oa-tools*.rpm") // Aggiunta ricerca RPM

	var allPackages []string
	allPackages = append(allPackages, debFiles...)
	allPackages = append(allPackages, archFiles...)
	allPackages = append(allPackages, rpmFiles...) // Inclusione nel pool di invio

	if len(allPackages) == 0 {
		fmt.Println("\033[1;31m[ERROR]\033[0m No native packages found.")
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
		fmt.Printf("\033[1;33m[CLEAN]\033[0m Removing old oa-tools packages on %s...\n", remoteUserHost)

		// 2. Costruzione dinamica del target di pulizia
		var rmTargets string
		if len(debFiles) > 0 {
			rmTargets += remotePkgPath + "oa-tools*.deb "
		}
		if len(archFiles) > 0 {
			rmTargets += remotePkgPath + "oa-tools*.pkg.tar.zst "
		}
		if len(rpmFiles) > 0 {
			rmTargets += remotePkgPath + "oa-tools*.rpm " // Aggiunta pulizia RPM
		}

		cleanCmdStr := "rm -f " + rmTargets
		sshArgs := append(muxArgs, remoteUserHost, cleanCmdStr)

		cleanCmd := exec.Command("ssh", sshArgs...)
		cleanCmd.Stdout, cleanCmd.Stderr, cleanCmd.Stdin = os.Stdout, os.Stderr, os.Stdin

		if err := cleanCmd.Run(); err != nil {
			fmt.Printf("\033[1;31m[ERROR]\033[0m Remote cleanup failed: %v\n", err)
		} else {
			fmt.Println("\033[1;32m[CLEAN]\033[0m Old packages removed.")
		}
	}

	// 3. Invio di tutti i pacchetti trovati
	for i, pkg := range allPackages {
		fmt.Printf("\033[1;34m[COPY]\033[0m Sending \033[1m%s\033[0m to Proxmox...\n", pkg)

		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remotePkgPath)
		scpArgs := append(muxArgs, pkg, dstStr)

		scpCmd := exec.Command("scp", scpArgs...)
		if i == 0 {
			scpCmd.Stdout, scpCmd.Stderr, scpCmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		} else {
			scpCmd.Stdout, scpCmd.Stderr = os.Stdout, os.Stderr
		}

		if err := scpCmd.Run(); err != nil {
			fmt.Printf("\033[1;31m[ERROR]\033[0m SCP transfer failed for %s: %v\n", pkg, err)
		} else {
			fmt.Printf("\033[1;32m[SUCCESS]\033[0m %s successfully exported to Proxmox.\n", pkg)
		}
	}
}
