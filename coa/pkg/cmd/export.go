package cmd

import (
	"github.com/spf13/cobra"
)

// --- CONFIGURAZIONE ESPORTAZIONE CONDIVISA ---
const (
	remoteUserHost = "root@192.168.1.2"
	remoteIsoPath  = "/var/lib/vz/template/iso/"
	remotePkgPath  = "/eggs/"
	isoSrcDir      = "/home/eggs"
)

// cleanExport è accessibile da export_iso.go e export_pkg.go
var cleanExport bool

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export artifacts (iso, pkg) to a remote Proxmox storage",
	Long: `Il comando export gestisce il trasferimento degli artefatti prodotti 
(ISO di installazione o pacchetti di distribuzione) verso server remoti 
configurati, automatizzando la pulizia e il versionamento.`,
}

func init() {
	// Aggiungiamo il flag 'clean' come persistente, così vale per tutti i sotto-comandi
	exportCmd.PersistentFlags().BoolVar(&cleanExport, "clean", false, "Clean old versions on remote server before exporting")
	rootCmd.AddCommand(exportCmd)
}
