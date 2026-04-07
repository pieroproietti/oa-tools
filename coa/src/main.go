// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	// 1. Discovery immediato dell'ambiente (Sensi)
	myDistro := NewDistro()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// 2. Gestione dei sotto-comandi (Logica)
	switch os.Args[1] {
	case "adapt":
		handleAdapt()
	case "produce":
		handleProduce(os.Args[2:], myDistro)
	case "export":
		handleExport(os.Args[2:]) // Nuovo comando
	case "kill":
		handleKill()
	case "detect":
		handleDetect(myDistro)
	case "version":
		fmt.Printf("coa v0.1.0 - The Mind of remaster\n")
	default:
		fmt.Printf("\033[1;31mError:\033[0m Unknown command '%s'\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// getOaPath cerca il braccio operativo (oa) nel sistema o nel percorso relativo
func getOaPath() string {
	path, err := exec.LookPath("oa")
	if err == nil {
		return path
	}
	if _, err := os.Stat("../oa/oa"); err == nil {
		return "../oa/oa"
	}
	if _, err := os.Stat("./oa/oa"); err == nil {
		return "./oa/oa"
	}
	return "oa"
}

// bridgeConfigs sovrascrive la configurazione mkinitcpio nella liveroot per Arch
func bridgeConfigs(d *Distro, workPath string) error {
	if d.FamilyID != "archlinux" {
		return nil
	}

	presetName := "live-arch.conf"
	if d.DistroID == "manjaro" || d.DistroID == "biglinux" {
		presetName = fmt.Sprintf("live-%s.conf", d.DistroID)
	}

	src := fmt.Sprintf("/tmp/coa/configs/mkinitcpio/%s", presetName)
	// Destinazione: Sovrascriviamo il file standard per evitare parametri extra nel chroot
	dst := filepath.Join(workPath, "liveroot", "etc", "mkinitcpio.conf")

	fmt.Printf("\033[1;34m[coa]\033[0m Overwriting liveroot /etc/mkinitcpio.conf with %s...\n", presetName)

	if err := copyFile(src, dst); err != nil {
		return err
	}

	return os.Chmod(dst, 0644)
}

// copyFile è una utility per la copia fisica tra percorsi host
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// handleProduce gestisce la creazione della ISO
func handleProduce(args []string, d *Distro) {
	produceCmd := flag.NewFlagSet("produce", flag.ExitOnError)
	mode := produceCmd.String("mode", "standard", "standard, clone, or crypted")
	workPath := produceCmd.String("path", "/home/eggs", "working directory")
	produceCmd.Parse(args)

	// 1. Estrazione Assets Config sull'host
	tempConfigPath := "/tmp/coa/configs"
	fmt.Printf("\033[1;32m[coa]\033[0m Extracting internal configurations to %s...\n", tempConfigPath)
	if err := ExtractConfigs(tempConfigPath); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Asset extraction failed: %v", err)
	}

	// 2. Download/Verifica Bootloaders (Indispensabile per action_iso)
	fmt.Printf("\033[1;32m[coa]\033[0m Ensuring bootloaders are present in %s...\n", BootloaderRoot)
	if _, err := EnsureBootloaders(); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Bootloader retrieval failed: %v", err)
	}

	// 3. Fase di Preparazione (Crea la liveroot tramite OverlayFS)
	fmt.Printf("\033[1;32m[coa]\033[0m Preparing environment...\n")
	prePlan := FlightPlan{
		PathLiveFs: *workPath,
		Mode:       *mode,
		Plan:       []Action{{Command: "action_prepare"}},
	}
	executePlan(prePlan)

	// 4. Ponte delle configurazioni (Host -> Liveroot)
	if err := bridgeConfigs(d, *workPath); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Config bridging failed: %v", err)
	}

	// 5. Esecuzione del piano di produzione completo
	fmt.Printf("\033[1;32m[coa]\033[0m Starting production flight...\n")
	flightPlan := GeneratePlan(d, *mode, *workPath)

	// Rimuoviamo action_prepare dal piano finale per evitare ri-montaggi
	if len(flightPlan.Plan) > 0 && flightPlan.Plan[0].Command == "action_prepare" {
		flightPlan.Plan = flightPlan.Plan[1:]
	}

	executePlan(flightPlan)
}

// handleKill gestisce la pulizia (ex eggs kill)
func handleKill() {
	fmt.Println("\033[1;33m[coa]\033[0m Freeing the nest...")

	// 1. Eseguiamo il cleanup chirurgico tramite oa (smontaggio)
	oaPath := getOaPath()
	// Passiamo un JSON minimale o usiamo il comando diretto se implementato
	cmd := exec.Command("sudo", oaPath, "cleanup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Cleanup (unmount) failed: %v\n", err)
	}

	// 2. Rimozione fisica della directory /home/eggs (il nest)
	workPath := "/home/eggs" // Default, o recuperalo dai flag
	fmt.Printf("\033[1;31m[coa]\033[0m Removing workspace: %s\n", workPath)

	// Usiamo rm -rf con cautela via sudo per pulire i file creati come root
	rmCmd := exec.Command("sudo", "rm", "-rf", workPath)
	if err := rmCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Physical removal failed: %v\n", err)
	} else {
		fmt.Println("\033[1;32m[coa]\033[0m Nest is empty. System clean.")
	}
}

// handleDetect mostra le info rilevate dal discovery
func handleDetect(d *Distro) {
	fmt.Println("\033[1;34m--- coa distro detect ---\033[0m")
	fmt.Printf("Host Distro:     %s\n", d.DistroID)
	fmt.Printf("Family:          %s\n", d.FamilyID)
	fmt.Printf("DistroLike:      %s\n", d.DistroLike)
	fmt.Printf("Codename:        %s\n", d.CodenameID)
	fmt.Printf("Release:         %s\n", d.ReleaseID)
	fmt.Printf("DistroUniqueID:  %s\n", d.DistroUniqueID)
}

// executePlan trasforma il piano in JSON e lo dà in pasto a oa
func executePlan(plan FlightPlan) {
	jsonData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m JSON Error: %v", err)
	}

	tmpJsonPath := "/tmp/plan_coa_tmp.json"
	err = os.WriteFile(tmpJsonPath, jsonData, 0644)
	if err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Temp file error: %v", err)
	}
	defer os.Remove(tmpJsonPath)

	oaPath := getOaPath()
	cmd := exec.Command("sudo", oaPath, tmpJsonPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("\n\033[1;31m[coa]\033[0m Engine error: %v", err)
	}
}

/**
*
 */
func printUsage() {
	fmt.Println("\033[1;32mcoa (Cova) - The Artisan Orchestrator\033[0m")
	fmt.Println("\nUsage:")
	fmt.Println("  coa produce [--mode standard|clone|crypted] [--path /path]")
	fmt.Println("  coa export [--dest /path] [--clean]") //
	fmt.Println("  coa kill")
	fmt.Println("  coa adapt")
	fmt.Println("  coa detect")
	fmt.Println("  coa version")
}

// handleAdapt adatta la risoluzione del monitor per le VM
// handleAdapt adatta la risoluzione del monitor per le VM (silenziosamente)
func handleAdapt() {
	fmt.Println("\033[1;33m[coa]\033[0m Adapting monitor resolution...")

	virtualOutputs := []string{"Virtual-0", "Virtual-1", "Virtual-2", "Virtual-3"}

	for _, output := range virtualOutputs {
		// Eseguiamo il comando senza collegare Stdout/Stderr
		cmd := exec.Command("xrandr", "--output", output, "--auto")

		// Run() attende la fine del comando.
		// Non avendo assegnato Stdout/Stderr, il terminale resta pulito.
		_ = cmd.Run()
	}

	fmt.Println("\033[1;32m[coa]\033[0m Resolution adapted.")
}

func handleExport(args []string) {
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	clean := exportCmd.Bool("clean", false, "remove previous versions")
	exportCmd.Parse(args)

	remoteHost := "root@192.168.1.2"
	remotePath := "/var/lib/vz/template/iso/"
	srcDir := "/home/eggs"

	// --- 1. SELEZIONE LOCALE: Identifichiamo l'uovo più fresco --- [cite: 29]
	allFiles, _ := filepath.Glob(filepath.Join(srcDir, "egg-of_*.iso"))
	if len(allFiles) == 0 {
		fmt.Println("\033[1;31m[coa]\033[0m Nest is empty.")
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

	// --- 2. MOUNT: Prepariamo il terreno --- [cite: 501, 502]
	localMount := "/tmp/coa-export-point"
	exec.Command("sudo", "fusermount", "-uz", localMount).Run()
	exec.Command("sudo", "rm", "-rf", localMount).Run()
	os.MkdirAll(localMount, 0755)

	fmt.Printf("\033[1;34m[coa]\033[0m Mounting Proxmox storage (root)...\n")
	// Importante: allow_other richiede user_allow_other in /etc/fuse.conf
	mountCmd := exec.Command("sshfs", remoteHost+":"+remotePath, localMount, "-o", "cache=no,allow_other")
	if out, err := mountCmd.CombinedOutput(); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Mount failed: %v\n%s\n", err, out)
		return
	}

	// Defer per garantire lo smontaggio e il sync finale [cite: 508, 509]
	defer func() {
		fmt.Printf("\033[1;34m[coa]\033[0m Finalizing: syncing and unmounting...\n")
		exec.Command("sync").Run()
		time.Sleep(1 * time.Second) // Piccolo respiro per FUSE
		exec.Command("sudo", "fusermount", "-uz", localMount).Run()
		exec.Command("sudo", "rm", "-rf", localMount).Run()
	}()

	// --- 3. PULIZIA E COPIA ---
	for prefix, localPath := range latestFiles {
		targetFileName := filepath.Base(localPath)
		fmt.Printf("\033[1;35m[PROCESS]\033[0m Family: %s\n", prefix)

		// CANCELLAZIONE: Scansione del server [cite: 504, 505]
		if *clean {
			remoteEntries, _ := os.ReadDir(localMount)
			for _, entry := range remoteEntries {
				// CANCELLA solo se inizia con lo stesso prefisso MA ha un nome diverso [cite: 509]
				if strings.HasPrefix(entry.Name(), prefix) && entry.Name() != targetFileName {
					fmt.Printf("\033[1;31m[DELETE]\033[0m Removing old version: %s\n", entry.Name())
					os.Remove(filepath.Join(localMount, entry.Name()))
				}
			}
		}

		// COPIA: Invio fisico del file [cite: 14, 570]
		dstPath := filepath.Join(localMount, targetFileName)
		fmt.Printf("\033[1;32m[COPY]\033[0m Sending %s to Proxmox...\n", targetFileName)

		if err := copyFile(localPath, dstPath); err != nil {
			fmt.Printf("\033[1;31m[ERROR]\033[0m Copy failed: %v\n", err)
		} else {
			fmt.Printf("\033[1;32m[SUCCESS]\033[0m %s is now on Proxmox.\n", targetFileName)
		}
	}
}
