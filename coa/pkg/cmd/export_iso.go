package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"coa/pkg/distro"

	"github.com/spf13/cobra"
)

var exportIsoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Export the latest ISO to a remote Proxmox storage",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
		handleExportIso(cleanExport) // Usa la variabile globale di export.go
	},
}

func init() {
	exportCmd.AddCommand(exportIsoCmd)
}

func handleExportIso(clean bool) {
	// 1. Otteniamo il prefisso dinamico (egg-of-distro-host-arch-)
	d := distro.NewDistro()
	prefixBase := d.GetISOPrefix()
	isoPattern := prefixBase + "*.iso"

	// Ricerca nel nido locale
	allFiles, _ := filepath.Glob(filepath.Join(isoSrcDir, isoPattern))
	if len(allFiles) == 0 {
		LogError("Il nido è vuoto per il prefisso: %s", prefixBase)
		return
	}

	// 2. Identificazione dell'ultima ISO (basata su ModTime)
	var latestFile string
	var latestTime time.Time

	for _, path := range allFiles {
		if info, err := os.Stat(path); err == nil {
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = path
			}
		}
	}

	targetFileName := filepath.Base(latestFile)
	LogCoala("Ultima ISO trovata: %s", targetFileName)

	// 3. Setup SSH Multiplexing
	socketPath := "/tmp/coa-ssh-mux"
	muxArgs := []string{"-o", "ControlPath=" + socketPath}

	// Start Master Connection
	exec.Command("ssh", "-M", "-f", "-N", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
	defer exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()

	// 4. Logica di Pulizia su Proxmox
	if clean {
		LogCoala("Pulizia su Proxmox: rimozione versioni precedenti con prefisso %s", prefixBase)
		rmCmdStr := fmt.Sprintf("rm -f %s/%s*.iso", remoteIsoPath, prefixBase)
		sshCmd := exec.Command("ssh", append(muxArgs, remoteUserHost, rmCmdStr)...)
		if err := sshCmd.Run(); err != nil {
			LogCoala("Nessuna vecchia ISO rimossa su Proxmox.")
		}
	}

	// 5. Invio effettivo
	LogCoala("Inviando %s verso Proxmox...", targetFileName)
	dst := fmt.Sprintf("%s:%s", remoteUserHost, remoteIsoPath)
	scpCmd := exec.Command("scp", append(muxArgs, latestFile, dst)...)
	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr

	if err := scpCmd.Run(); err != nil {
		LogError("Errore durante il trasferimento: %v", err)
	} else {
		LogSuccess("Esportazione completata con successo!")
	}
}
