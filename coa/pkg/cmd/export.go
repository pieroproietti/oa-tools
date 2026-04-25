package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"coa/pkg/distro" // Assicurati che il path sia corretto per il tuo progetto

	"github.com/spf13/cobra"
)

// --- CONFIGURAZIONE ESPORTAZIONE ---
const (
	remoteUserHost = "root@192.168.1.2"
	remoteIsoPath  = "/var/lib/vz/template/iso/"
	remotePkgPath  = "/eggs/"
	isoSrcDir      = "/home/eggs"
)

var cleanExport bool

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export artifacts (iso, pkg) to a remote Proxmox storage",
}

var exportIsoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Export the latest ISO to a remote Proxmox storage",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
		handleExportIso(cleanExport)
	},
}

var exportPkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Export native packages (.deb, .rpm, .pkg.tar.zst) to Proxmox",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
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
// LOGICA DI ESPORTAZIONE
// =====================================================================

// handleExportPkg esporta solo i pacchetti della distro corrente
func handleExportPkg(clean bool) {
	myDistro := distro.NewDistro()
	family := myDistro.FamilyID

	LogCoala("Famiglia rilevata: %s. Ricerca pacchetti pertinenti...", family)

	var pattern string
	var extension string

	// Filtriamo per estensione in base alla famiglia
	switch family {
	case "debian", "ubuntu", "devuan":
		pattern = "oa-tools*.deb"
		extension = ".deb"
	case "arch":
		pattern = "oa-tools*.pkg.tar.zst"
		extension = ".pkg.tar.zst"
	case "fedora", "redhat", "suse":
		pattern = "oa-tools*.rpm"
		extension = ".rpm"
	default:
		// Se la famiglia non è riconosciuta, usiamo LogCoala per avvisare
		LogCoala("Nessuna regola di esportazione specifica per la famiglia: %s", family)
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

// handleExportIso (Logica classica con ridenominazione post-processo)
func handleExportIso(clean bool) {
	// 1. Ricerca dei file nel nido con il prefisso classico
	allFiles, _ := filepath.Glob(filepath.Join(isoSrcDir, "egg-of_*.iso"))
	if len(allFiles) == 0 {
		LogError("Il nido è vuoto. Nessuna ISO trovata in %s con prefisso 'egg-of_'", isoSrcDir)
		return
	}

	// 2. Logica per identificare solo l'ultima versione prodotta per ogni famiglia
	latestFiles := make(map[string]string)
	// Regex per isolare il timestamp (es. _2026-04-25_1200.iso)
	re := regexp.MustCompile(`_\d{4}-\d{2}-\d{2}_\d{4}\.iso$`)

	for _, path := range allFiles {
		fileName := filepath.Base(path)
		// Otteniamo il prefisso rimuovendo la parte del tempo (es. egg-of_debian-bookworm)
		prefix := re.ReplaceAllString(fileName, "")

		if info, err := os.Stat(path); err == nil {
			if current, exists := latestFiles[prefix]; exists {
				cInfo, _ := os.Stat(current)
				// Se il file attuale è più recente di quello salvato, lo sostituiamo
				if info.ModTime().After(cInfo.ModTime()) {
					latestFiles[prefix] = path
				}
			} else {
				latestFiles[prefix] = path
			}
		}
	}

	LogCoala("ISO identificate. Preparazione del tunnel SSH...")

	// 3. Setup SSH Multiplexing per velocizzare l'invio
	socketPath := "/tmp/coa-ssh-mux"
	muxArgs := []string{
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=" + socketPath,
		"-o", "ControlPersist=2m",
	}

	// Assicuriamoci di chiudere il socket alla fine, qualunque cosa accada
	defer func() {
		exec.Command("ssh", "-O", "exit", "-o", "ControlPath="+socketPath, remoteUserHost).Run()
		os.Remove(socketPath)
	}()

	// 4. Ciclo di esportazione per ogni famiglia trovata
	for prefix, localPath := range latestFiles {
		targetFileName := filepath.Base(localPath)
		LogCoala("Processando famiglia: %s", prefix)

		// Se richiesto, puliamo le vecchie versioni sul server Proxmox
		if clean {
			LogCoala("Pulizia vecchie versioni su Proxmox per %s...", prefix)
			rmCmdStr := fmt.Sprintf("rm -f %s%s*", remoteIsoPath, prefix)
			sshArgs := append(muxArgs, remoteUserHost, rmCmdStr)

			if err := exec.Command("ssh", sshArgs...).Run(); err != nil {
				LogCoala("Nessun vecchio file rimosso (pulizia non necessaria o fallita).")
			} else {
				LogSuccess("Vecchie versioni rimosse dal server.")
			}
		}

		// Invio effettivo tramite SCP
		LogCoala("Invio di %s a Proxmox...", targetFileName)
		dstStr := fmt.Sprintf("%s:%s", remoteUserHost, remoteIsoPath)
		scpArgs := append(muxArgs, localPath, dstStr)

		scpCmd := exec.Command("scp", scpArgs...)
		// Agganciamo l'output per vedere il progresso di SCP
		scpCmd.Stdout = os.Stdout
		scpCmd.Stderr = os.Stderr

		if err := scpCmd.Run(); err != nil {
			LogError("Trasferimento fallito per %s: %v", targetFileName, err)
		} else {
			LogSuccess("%s esportata correttamente su Proxmox.", targetFileName)
		}
	}
}
