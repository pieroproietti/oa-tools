// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package krill

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	// Importiamo l'engine per usare le sue strutture dati (FlightPlan, Action, ecc.)
	"coa/src/internal/engine"
)

// KrillAnswers raccoglie i dati inseriti dall'utente nella TUI
type KrillAnswers struct {
	TargetDisk string
	Hostname   string
	Username   string
	Password   string
	UseLuks    bool
}

// HandleKrill avvia l'interfaccia interattiva
func HandleKrill() {
	fmt.Println("\033[1;36m====================================================\033[0m")
	fmt.Println("\033[1;36m            KRILL - The coa System Installer        \033[0m")
	fmt.Println("\033[1;36m====================================================\033[0m")

	disks := getAvailableDisks()
	answers := &KrillAnswers{}

	qs := []*survey.Question{
		{
			Name: "TargetDisk",
			Prompt: &survey.Select{
				Message: "Select the target disk for installation (WARNING: Will be wiped!):",
				Options: disks,
			},
			Validate: survey.Required,
		},
		{
			Name: "Hostname",
			Prompt: &survey.Input{
				Message: "Enter the new Hostname:",
				Default: "coa-machine",
			},
			Validate: survey.Required,
		},
		{
			Name: "Username",
			Prompt: &survey.Input{
				Message: "Enter the new login Username:",
				Default: "artisan",
			},
			Validate: survey.Required,
		},
		{
			Name: "Password",
			Prompt: &survey.Password{
				Message: "Enter User Password:",
			},
			Validate: survey.Required,
		},
		{
			Name: "UseLuks",
			Prompt: &survey.Confirm{
				Message: "Do you want to encrypt the entire system with LUKS2?",
				Default: false,
			},
		},
	}

	if err := survey.Ask(qs, answers); err != nil {
		fmt.Printf("\033[1;31mInstallation aborted:\033[0m %v\n", err)
		return
	}

	if answers.TargetDisk == "NO_SAFE_DISKS_FOUND" {
		fmt.Println("\033[1;31mNo safe target disks available. Aborting.\033[0m")
		return
	}

	targetClean := strings.Split(answers.TargetDisk, " ")[0]

	fmt.Println("\n\033[1;33m--- Installation Summary ---\033[0m")
	fmt.Printf("Disk:      \033[1;31m%s (Will be WIPED)\033[0m\n", targetClean)
	fmt.Printf("Hostname:  %s\n", answers.Hostname)
	fmt.Printf("Username:  %s\n", answers.Username)
	fmt.Printf("LUKS:      %v\n", answers.UseLuks)

	confirm := false
	survey.AskOne(&survey.Confirm{
		Message: "Are you absolutely sure you want to proceed? This is destructive.",
		Default: false,
	}, &confirm)

	if !confirm {
		fmt.Println("\033[1;32mInstallation canceled. Nest is safe.\033[0m")
		return
	}

	generateInstallPlan(answers, targetClean)
}

func generateInstallPlan(ans *KrillAnswers, disk string) {
	fmt.Println("\n\033[1;34m[krill]\033[0m Compiling JSON flight plan for the C engine...")

	squashPath := getSquashfsPath()
	if squashPath == "" {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Could not locate pristine filesystem.squashfs on this live system!\n")
		return
	}
	fmt.Printf("\033[1;34m[krill]\033[0m Pristine image located at: %s\n", squashPath)

	hashedPass := generateHashedPassword(ans.Password)

	// Riutilizziamo la struttura dell'engine
	plan := engine.FlightPlan{
		PathLiveFs: "/mnt/krill-target",
		Mode:       "install",
		Plan: []engine.Action{
			{
				Command:    "oa_install_partition",
				RunCommand: disk,
			},
		},
	}

	if ans.UseLuks {
		plan.Plan = append(plan.Plan, engine.Action{
			Command:         "oa_install_format_luks",
			RunCommand:      disk,
			CryptedPassword: ans.Password,
		})
	} else {
		plan.Plan = append(plan.Plan, engine.Action{
			Command:    "oa_install_format",
			RunCommand: disk,
		})
	}

	plan.Plan = append(plan.Plan,
		engine.Action{
			Command:    "oa_install_unpack",
			RunCommand: disk,
			Args:       []string{squashPath},
		},
		engine.Action{Command: "oa_install_fstab", RunCommand: disk},
		engine.Action{
			Command:    "oa_sys_run",
			RunCommand: "sh",
			Args:       []string{"-c", "rm -f /etc/machine-id && touch /etc/machine-id"},
		},
		engine.Action{
			Command:    "oa_sys_run",
			RunCommand: "sh",
			Args:       []string{"-c", "echo " + ans.Hostname + " > /etc/hostname"},
		},
		engine.Action{Command: "oa_install_users"},
		engine.Action{Command: "oa_install_initrd"},
	)

	if isUEFI() {
		fmt.Println("\033[1;34m[krill]\033[0m UEFI boot detected. Planning EFI Grub installation.")
		plan.Plan = append(plan.Plan, engine.Action{Command: "oa_install_uefi", RunCommand: disk})
	} else {
		fmt.Println("\033[1;34m[krill]\033[0m Legacy BIOS boot detected. Planning PC Grub installation.")
		plan.Plan = append(plan.Plan, engine.Action{Command: "oa_install_bios", RunCommand: disk})
	}

	plan.Users = []engine.UserConfig{
		{
			Login:    ans.Username,
			Password: hashedPass,
			Gecos:    ans.Username + ",,,",
			Home:     "/home/" + ans.Username,
			Shell:    "/bin/bash",
			Groups:   []string{"sudo", "wheel", "video", "audio", "plugdev", "netdev"},
		},
	}

	jsonData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Failed to marshal JSON: %v\n", err)
		return
	}

	outPath := "/tmp/sysinstall.json"
	if err := os.WriteFile(outPath, jsonData, 0644); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Failed to write plan: %v\n", err)
		return
	}

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Flight plan compiled and saved to \033[1m%s\033[0m\n", outPath)
	fmt.Println("\033[1;33m[krill]\033[0m To execute physical installation: \033[1msudo oa /tmp/plan-install.json\033[0m")

	// Chiamata all'engine centrale
	engine.ExecutePlan(plan)
}

// =========================================================================
// HELPER DI SISTEMA (invariati rispetto al tuo codice originale)
// =========================================================================

func isUEFI() bool {
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		return true
	}
	return false
}

// getSquashfsPath
func getSquashfsPath() string {
	paths := []string{
		// Percorsi Debian/Ubuntu Live
		"/run/live/medium/live/filesystem.squashfs",
		"/lib/live/mount/medium/live/filesystem.squashfs",

		// Percorso Arch Linux (archiso)
		"/run/archiso/bootmnt/arch/x86_64/airootfs.sfs",

		// Percorso Remastering locale coa/oa
		"/home/eggs/iso/live/filesystem.squashfs",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func getRootDisk() string {
	cmd := "lsblk -n -o PKNAME $(findmnt -n -v -o SOURCE /) | head -n 1"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getAvailableDisks() []string {
	rootDisk := getRootDisk()

	out, err := exec.Command("lsblk", "-d", "-n", "-o", "NAME,SIZE,MODEL").Output()
	if err != nil || len(out) == 0 {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var disks []string

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			diskName := parts[0]

			if diskName == rootDisk {
				continue
			}

			if strings.HasPrefix(diskName, "loop") || strings.HasPrefix(diskName, "sr") {
				continue
			}

			diskStr := fmt.Sprintf("/dev/%s - %s %s", diskName, parts[1], strings.Join(parts[2:], " "))
			disks = append(disks, diskStr)
		}
	}

	if len(disks) == 0 {
		disks = append(disks, "NO_SAFE_DISKS_FOUND")
	}

	return disks
}

func generateHashedPassword(plain string) string {
	out, err := exec.Command("openssl", "passwd", "-6", plain).Output()
	if err != nil {
		return plain
	}
	return strings.TrimSpace(string(out))
}
