package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

// Flags
var cleanExport bool

// Parent Command: export
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export artifacts (iso, pkg) to a remote Proxmox storage",
	Long:  "Export generated ISOs or native packages to a remote server via SCP.",
}

// Subcommand: iso
var exportIsoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Export the latest ISO to a remote Proxmox storage",
	Run: func(cmd *cobra.Command, args []string) {
		// INSERISCI QUI LA TUA LOGICA ESISTENTE PER LA ISO
		fmt.Println("\033[1;34m[PROCESS]\033[0m Starting ISO export...")
		// ...
	},
}

// Subcommand: pkg
var exportPkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Export the latest generated native package (.deb or .pkg.tar.zst) to Proxmox",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\033[1;34m[PROCESS]\033[0m Searching for native packages...")

		// 1. Find packages for both Debian and Arch Linux
		debFiles, _ := filepath.Glob("*.deb")
		archFiles, _ := filepath.Glob("*.pkg.tar.zst")

		var allPackages []string
		allPackages = append(allPackages, debFiles...)
		allPackages = append(allPackages, archFiles...)

		if len(allPackages) == 0 {
			fmt.Println("\033[1;31m[ERROR]\033[0m No native packages (.deb or .pkg.tar.zst) found in the current directory.")
			return
		}

		// 2. Sort by modification time (newest last)
		sort.Slice(allPackages, func(i, j int) bool {
			infoI, _ := os.Stat(allPackages[i])
			infoJ, _ := os.Stat(allPackages[j])
			return infoI.ModTime().Before(infoJ.ModTime())
		})

		latestPkg := allPackages[len(allPackages)-1]
		targetIP := "192.168.1.2"
		targetPath := "/eggs/"

		// 3. Optional remote cleanup
		if cleanExport {
			fmt.Printf("\033[1;33m[CLEAN]\033[0m Removing old packages on %s...\n", targetIP)
			// Rimuove sia i vecchi deb che i vecchi pacchetti arch
			cleanCmd := exec.Command("ssh", "root@"+targetIP, "rm -f "+targetPath+"*.deb "+targetPath+"*.pkg.tar.zst")
			cleanCmd.Stdout = os.Stdout
			cleanCmd.Stderr = os.Stderr
			cleanCmd.Run()
			fmt.Println("\033[1;32m[CLEAN]\033[0m Old packages removed.")
		}

		// 4. Execute SCP transfer
		fmt.Printf("\033[1;34m[COPY]\033[0m Sending \033[1m%s\033[0m to Proxmox...\n", latestPkg)

		scpCmd := exec.Command("scp", latestPkg, "root@"+targetIP+":"+targetPath)
		scpCmd.Stdout = os.Stdout
		scpCmd.Stderr = os.Stderr

		if err := scpCmd.Run(); err != nil {
			fmt.Printf("\033[1;31m[ERROR]\033[0m SCP transfer failed: %v\n", err)
			return
		}

		fmt.Printf("\033[1;32m[SUCCESS]\033[0m %s successfully exported to Proxmox.\n", latestPkg)
	},
}

// init registers the commands and flags
func init() {
	exportIsoCmd.Flags().BoolVar(&cleanExport, "clean", false, "Clean old versions on remote server before exporting")
	exportPkgCmd.Flags().BoolVar(&cleanExport, "clean", false, "Clean old packages on remote server before exporting")

	exportCmd.AddCommand(exportIsoCmd)
	exportCmd.AddCommand(exportPkgCmd)
}
