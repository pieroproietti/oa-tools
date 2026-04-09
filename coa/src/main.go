// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Variabili globali per i flag di Cobra
var (
	modeFlag  string
	pathFlag  string
	cleanFlag bool
)

func main() {
	// Discovery immediato dell'ambiente (Sensi)
	myDistro := NewDistro()

	// --- ROOT COMMAND ---
	var rootCmd = &cobra.Command{
		Use:   "coa",
		Short: "coa (Cova) - The Artisan Orchestrator",
		Long:  "coa is the intelligent orchestrator written in Go, designed to be the \"Mind\" behind the GNU/Linux system remastering process.",
	}

	// --- PRODUCE COMMAND ---
	var produceCmd = &cobra.Command{
		Use:   "produce",
		Short: "Start a system remastering production flight",
		Run: func(cmd *cobra.Command, args []string) {
			handleProduce(modeFlag, pathFlag, myDistro)
		},
	}
	produceCmd.Flags().StringVar(&modeFlag, "mode", "standard", "standard, clone, or crypted")
	produceCmd.Flags().StringVar(&pathFlag, "path", "/home/eggs", "working directory")

	// --- EXPORT COMMAND ---
	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export the latest ISO to a remote Proxmox storage",
		Run: func(cmd *cobra.Command, args []string) {
			handleExport(cleanFlag)
		},
	}
	exportCmd.Flags().BoolVar(&cleanFlag, "clean", false, "remove previous versions")

	// --- KILL COMMAND ---
	var killCmd = &cobra.Command{
		Use:   "kill",
		Short: "Free the nest and unmount filesystems",
		Run: func(cmd *cobra.Command, args []string) {
			handleKill()
		},
	}

	// --- DETECT COMMAND ---
	var detectCmd = &cobra.Command{
		Use:   "detect",
		Short: "Show host distribution discovery info",
		Run: func(cmd *cobra.Command, args []string) {
			handleDetect(myDistro)
		},
	}

	// --- KRILL COMMAND ---
	var krillCmd = &cobra.Command{
		Use:   "krill",
		Short: "Start the system installation (The Hatching)",
		Run: func(cmd *cobra.Command, args []string) {
			handleKrill()
		},
	}

	// --- ADAPT COMMAND ---
	var adaptCmd = &cobra.Command{
		Use:   "adapt",
		Short: "Adapt monitor resolution for VMs",
		Run: func(cmd *cobra.Command, args []string) {
			handleAdapt()
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of coa",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("coa v%s - The Mind of remaster\n", AppVersion)
		},
	}

	// --- DOCS COMMAND (GENERATORE WIKI, MAN E AUTOCOMPLETE) ---
	var docsCmd = &cobra.Command{
		Use:    "docs",
		Short:  "Generate man pages, markdown wiki, and completion scripts",
		Hidden: true, // Nascondiamo il comando dall'help principale
		Run: func(cmd *cobra.Command, args []string) {
			generateDocs(rootCmd)
		},
	}

	// --- BUILD COMMAND ---
	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Compile binaries and generate native distribution packages (.deb, PKGBUILD)",
		Run: func(cmd *cobra.Command, args []string) {
			handleBuild(myDistro)
		},
	}

	// Assicurati di aggiungere buildCmd alla riga rootCmd.AddCommand in fondo al file:
	rootCmd.AddCommand(
		adaptCmd,
		buildCmd,
		detectCmd,
		docsCmd,
		exportCmd,
		killCmd,
		krillCmd,
		produceCmd,
		versionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
