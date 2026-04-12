package cmd

import (
	"coa/src/internal/engine"

	"github.com/spf13/cobra"
)

var cleanExport bool

var exportCmd = &cobra.Command{
	Use:    "export",
	Short:  "Export artifacts (iso, pkg) to a remote Proxmox storage",
	Long:   "Export generated ISOs or native packages to a remote server via SCP.",
	Hidden: false, // Sparisce da coa --help, ma continua a funzionare!
}

var exportIsoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Export the latest ISO to a remote Proxmox storage",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements("export iso", false)
		engine.HandleExportIso(cleanExport)
	},
}

var exportPkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Export the latest generated native package (.deb or .pkg.tar.zst) to Proxmox",
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements("export pkg", false)
		engine.HandleExportPkg(cleanExport)
	},
}

func init() {
	exportCmd.PersistentFlags().BoolVar(&cleanExport, "clean", false, "Clean old versions on remote server before exporting")

	exportCmd.AddCommand(exportIsoCmd)
	exportCmd.AddCommand(exportPkgCmd)
	rootCmd.AddCommand(exportCmd)
}
