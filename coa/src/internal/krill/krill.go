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
	fmt.Println("\033[1;36m       KRILL - The coa Universal Installer          \033[0m")
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
		fmt.Printf("\033[1;31m[ERROR]\033[0m Could not locate pristine filesystem.squashfs!\n")
		return
	}

	hashedPass := generateHashedPassword(ans.Password)

	// Inizializzazione del Piano di Volo
	plan := engine.FlightPlan{
		PathLiveFs: "/mnt/krill-target", // Area di mount per l'installazione
		Mode:       "install",
		Plan:       []engine.Action{},
	}

	// 1. DISCO E PARTIZIONAMENTO
	plan.Plan = append(plan.Plan, engine.Action{
		Command:    "oa_install_partition",
		RunCommand: disk,
	})

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

	// 2. UNPACK (Scompatta l'immagine e monta le API FS nel target)
	plan.Plan = append(plan.Plan, engine.Action{
		Command:    "oa_install_unpack",
		RunCommand: disk,
		Args:       []string{squashPath},
	})

	// 3. CONFIGURAZIONE CHROOT (Usando il nuovo oa_sys_shell universale)
	plan.Plan = append(plan.Plan,
		engine.Action{Command: "oa_install_fstab", RunCommand: disk},

		// Generazione Machine-ID univoco (Essenziale per systemd e networking)
		engine.Action{
			Command:    "oa_sys_shell",
			RunCommand: "rm -f /etc/machine-id /var/lib/dbus/machine-id && systemd-machine-id-setup || touch /etc/machine-id",
			Chroot:     true,
		},

		// Setup Hostname e aggiornamento /etc/hosts (per evitare lag di sudo)
		engine.Action{
			Command: "oa_sys_shell",
			RunCommand: fmt.Sprintf(
				"echo %s > /etc/hostname && sed -i 's/127.0.1.1.*/127.0.1.1\t%s/' /etc/hosts",
				ans.Hostname, ans.Hostname,
			),
			Chroot: true,
		},

		// Pulizia file residui della sessione Live
		engine.Action{
			Command:    "oa_sys_shell",
			RunCommand: "rm -rf /var/log/installer /var/lib/live/config /etc/sudoers.d/live-user 2>/dev/null",
			Chroot:     true,
		},

		engine.Action{Command: "oa_install_users"},
		engine.Action{Command: "oa_install_initrd"},
	)

	// 4. BOOTLOADER (UEFI vs BIOS)
	if isUEFI() {
		plan.Plan = append(plan.Plan, engine.Action{Command: "oa_install_uefi", RunCommand: disk})
	} else {
		plan.Plan = append(plan.Plan, engine.Action{Command: "oa_install_bios", RunCommand: disk})
	}

	// Configurazione Utente primario
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

	// Salvataggio ed Esecuzione
	jsonData, _ := json.MarshalIndent(plan, "", "  ")
	outPath := "/tmp/sysinstall.json"
	os.WriteFile(outPath, jsonData, 0644)

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Flight plan ready at %s\n", outPath)

	// Esecuzione tramite l'engine centrale
	engine.ExecutePlan(plan)
}

// --- HELPER DI SISTEMA ---

func isUEFI() bool {
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		return true
	}
	return false
}

func getSquashfsPath() string {
	paths := []string{
		"/run/live/medium/live/filesystem.squashfs",       // Debian Live
		"/lib/live/mount/medium/live/filesystem.squashfs", // Debian Alt
		"/run/archiso/bootmnt/arch/x86_64/airootfs.sfs",   // Arch
		"/home/eggs/iso/live/filesystem.squashfs",         // Local coa/oa
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
	out, _ := exec.Command("lsblk", "-d", "-n", "-o", "NAME,SIZE,MODEL").Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var disks []string

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			diskName := parts[0]
			if diskName == rootDisk || strings.HasPrefix(diskName, "loop") || strings.HasPrefix(diskName, "sr") {
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
