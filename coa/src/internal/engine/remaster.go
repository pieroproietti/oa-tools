package engine

import (
	"coa/src/internal/assets"
	"coa/src/internal/distro"
	"coa/src/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// HandleRemaster gestisce la creazione della ISO
func HandleRemaster(mode string, workPath string, d *distro.Distro) {
	tempConfigPath := "/tmp/coa/configs"
	fmt.Printf("\033[1;32m[coa]\033[0m Extracting internal configurations to %s...\n", tempConfigPath)

	// TODO: Chiamerà assets.ExtractConfigs(tempConfigPath)
	if err := assets.ExtractConfigs(tempConfigPath); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Asset extraction failed: %v", err)
	}

	fmt.Printf("\033[1;32m[coa]\033[0m Ensuring bootloaders are present...\n")

	// Passiamo la costante BootloaderRoot definita in plan.go
	if _, err := utils.EnsureBootloaders(BootloaderRoot); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Bootloader retrieval failed: %v", err)
	}

	fmt.Printf("\033[1;32m[coa]\033[0m Preparing environment...\n")

	// La struttura Action e FlightPlan andrà definita in un file plan.go qui in engine
	prePlan := FlightPlan{
		PathLiveFs: workPath,
		Mode:       mode,
		Plan:       []Action{{Command: "oa_remaster_prepare"}},
	}
	ExecutePlan(prePlan)

	if err := bridgeConfigs(d, workPath); err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m Config bridging failed: %v", err)
	}

	fmt.Printf("\033[1;32m[coa]\033[0m Starting production flight...\n")

	// TODO: Chiamerà GeneratePlan (da spostare in engine/plan.go)
	flightPlan := GeneratePlan(d, mode, workPath)

	if len(flightPlan.Plan) > 0 && flightPlan.Plan[0].Command == "action_prepare" {
		flightPlan.Plan = flightPlan.Plan[1:]
	}
	ExecutePlan(flightPlan)
}

// ExecutePlan trasforma il piano in JSON e lo dà in pasto a oa
func ExecutePlan(plan FlightPlan) {
	jsonData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		log.Fatalf("\033[1;31m[coa]\033[0m JSON Error: %v", err)
	}

	tmpJsonPath := "/tmp/remaster.json"
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
func bridgeConfigs(d *distro.Distro, workPath string) error {
	if d.FamilyID != "archlinux" {
		return nil
	}

	presetName := "live-arch.conf"
	if d.DistroID == "manjaro" || d.DistroID == "biglinux" {
		presetName = fmt.Sprintf("live-%s.conf", d.DistroID)
	}
	src := fmt.Sprintf("/tmp/coa/configs/mkinitcpio/%s", presetName)
	dst := filepath.Join(workPath, "liveroot", "etc", "coa_mkinitcpio.conf")

	fmt.Printf("\033[1;34m[coa]\033[0m Creating liveroot /etc/mkinitcpio.conf with %s...\n", presetName)

	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Chmod(dst, 0644)
}

// copyFile utility per la copia fisica tra percorsi host
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
