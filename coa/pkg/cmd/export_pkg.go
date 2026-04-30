package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"coa/pkg/distro"

	"github.com/spf13/cobra"
)

var exportPkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Export native packages (.deb, .rpm, .pkg.tar.zst) to Proxmox",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
		handleExportPkg(cleanExport) // Usa la variabile globale di export.go
	},
}

func init() {
	exportCmd.AddCommand(exportPkgCmd)
}

func handleExportPkg(clean bool) {
	myDistro := distro.NewDistro()
	distro:= myDistro.DistroID
	family := myDistro.FamilyID

	LogCoala("Famiglia rilevata: %s. Ricerca pacchetti pertinenti...", family)

	var pattern string
	var extension string

	// Filtriamo per estensione in base alla famiglia
	switch family {
	case "debian":
		pattern = "oa-tools*.deb"
		extension = ".deb"
	case "archlinux":
		pattern = "oa-tools*.pkg.tar.zst"
		extension = ".pkg.tar.zst"
	case "fedora":
		pattern = "oa-tools*.rpm"
		extension = ".rpm"
	case "manjaro":
		pattern = "oa-tools-manjaro-*.pkg.tar.zst"
		extension = ".pkg.tar.zst"
	default:
		LogCoala("Nessuna regola di esportazione specifica per la distro %s della famiglia: %s", distro,family)
		return
	}

	foundFiles, _ := filepath.Glob(pattern)
	if len(foundFiles) == 0 {
		LogError("Nessun pacchetto %s trovato per l'esportazione.", extension)
		return
	}

	// SSH Multiplexing
	socketPath := "/tmp/coa-ssh-mux-pkg"
	muxArgs := []string{"-o", "ControlMaster=auto", "-o", "ControlPath=" + socketPath, "-o", "ControlPersist=2m"}
	defer func() {
		exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
		os.Remove(socketPath)
	}()

	if clean {
		LogCoala("Pulizia remota vecchi pacchetti %s...", extension)
		cleanCmdStr := fmt.Sprintf("rm -f %soa-tools*%s", remotePkgPath, extension)
		sshArgs := append(muxArgs, remoteUserHost, cleanCmdStr)

		if err := exec.Command("ssh", sshArgs...).Run(); err != nil {
			LogCoala("Pulizia remota non necessaria o fallita (nessun file trovato).")
		} else {
			LogSuccess("Vecchi pacchetti %s rimossi dal server.", extension)
		}
	}

	for _, pkg := range foundFiles {
		LogCoala("Esportazione: %s", pkg)
		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remotePkgPath)
		scpArgs := append(muxArgs, pkg, dstStr)

		scpCmd := exec.Command("scp", scpArgs...)
		scpCmd.Stdout, scpCmd.Stderr = os.Stdout, os.Stderr

		if err := scpCmd.Run(); err != nil {
			LogError("Trasferimento fallito per %s: %v", pkg, err)
		} else {
			LogSuccess("%s inviato con successo.", pkg)
		}
	}
}
