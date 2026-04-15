// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"coa/src/internal/distro"
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
func HandleKrill(d *distro.Distro) {
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

	generateInstallPlan(answers, targetClean, d)
}

func generateInstallPlan(ans *KrillAnswers, disk string, d *distro.Distro) {
	fmt.Println("\n\033[1;34m[krill]\033[0m Compiling JSON flight plan for the C ..")

	squashPath := getSquashfsPath()
	if squashPath == "" {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Could not locate pristine filesystem.squashfs!\n")
		return
	}

	hashedPass := generateHashedPassword(ans.Password)

	// 1. Inizializzazione del Piano di Volo
	plan := FlightPlan{
		PathLiveFs: "/mnt/krill-target",
		Mode:       "install",
		Plan:       []Action{},
	}

	// 2. PREPARAZIONE DISCO E FILESYSTEM
	plan.Plan = append(plan.Plan, Action{Command: "oa_install_partition", RunCommand: disk})

	// FIX PER FEDORA 43: Formattazione EXT4 con flag di compatibilità per GRUB
	if d.FamilyID == "fedora" || d.FamilyID == "rhel" {
		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_shell",
			RunCommand: fmt.Sprintf("mkfs.fat -F32 %s2 && mkfs.ext4 -O ^metadata_csum_seed,^orphan_file %s3", disk, disk),
		})
	} else {
		plan.Plan = append(plan.Plan, Action{Command: "oa_install_format", RunCommand: disk})
	}

	plan.Plan = append(plan.Plan,
		Action{Command: "oa_install_unpack", RunCommand: disk, Args: []string{squashPath}},
		Action{Command: "oa_install_prepare", RunCommand: disk},
	)

	// 3. LOGICA DISTRO-SPECIFICA (The Mind)
	var shellInitrdCmd string
	var shellGrubCmd string

	switch d.FamilyID {
	case "archlinux":
		shellInitrdCmd = "mkinitcpio -P"
		if isUEFI() {
			shellGrubCmd = "grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=coa --recheck && grub-mkconfig -o /boot/grub/grub.cfg"
		} else {
			shellGrubCmd = "grub-install --target=i386-pc --recheck " + disk + " && grub-mkconfig -o /boot/grub/grub.cfg"
		}
	case "fedora", "rhel", "centos", "rocky", "almalinux":
		shellInitrdCmd = "dracut --force --regenerate-all"
		// FIX PER FEDORA: Disabilitazione BLS e installazione fisica GRUB
		if isUEFI() {
			shellGrubCmd = "echo 'GRUB_ENABLE_BLSCFG=false' >> /etc/default/grub && " +
				"grub2-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=fedora --recheck && " +
				"grub2-mkconfig -o /boot/grub2/grub.cfg && " +
				"cp /boot/grub2/grub.cfg /boot/efi/EFI/fedora/grub.cfg"
		} else {
			shellGrubCmd = "echo 'GRUB_ENABLE_BLSCFG=false' >> /etc/default/grub && " +
				"grub2-install --target=i386-pc --recheck " + disk + " && " +
				"grub2-mkconfig -o /boot/grub2/grub.cfg"
		}
	default: // Debian/Ubuntu
		shellInitrdCmd = "update-initramfs -u -k all"
		if isUEFI() {
			shellGrubCmd = "grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=coa --recheck && update-grub"
		} else {
			shellGrubCmd = "grub-install --target=i386-pc --recheck " + disk + " && update-grub"
		}
	}

	// 4. ESECUZIONE AZIONI UNIVERSALI E CHROOT
	plan.Plan = append(plan.Plan,
		Action{Command: "oa_install_fstab", RunCommand: disk},
		Action{
			Command:    "oa_sys_shell",
			RunCommand: fmt.Sprintf("echo %s > /etc/hostname && systemd-machine-id-setup", ans.Hostname),
			Chroot:     true,
		},
		Action{
			Command:    "oa_sys_shell",
			RunCommand: shellInitrdCmd,
			Chroot:     true,
		},
		Action{
			Command:    "oa_sys_shell",
			RunCommand: shellGrubCmd,
			Chroot:     true,
		},
		Action{Command: "oa_install_users"},
	)

	// 5. PULIZIA SELETTIVA E SYNC (FIX PER HANG SU FEDORA)
	if d.FamilyID == "debian" {
		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_shell",
			RunCommand: "rm -rf /var/log/installer /var/lib/live/config /etc/sudoers.d/live-user 2>/dev/null",
			Chroot:     true,
		})
	}

	plan.Plan = append(plan.Plan,
		Action{Command: "oa_sys_shell", RunCommand: "sync"},
		Action{Command: "oa_install_cleanup"},
	)

	// 6. Configurazione Utente Primario
	plan.Users = []UserConfig{
		{
			Login:    ans.Username,
			Password: hashedPass,
			Gecos:    ans.Username + ",,,",
			Home:     "/home/" + ans.Username,
			Shell:    "/bin/bash",
			Groups:   []string{"sudo", "wheel", "video", "audio", "plugdev", "netdev"},
		},
	}

	// 7. Salvataggio ed Esecuzione
	jsonData, _ := json.MarshalIndent(plan, "", "  ")
	outPath := "/tmp/sysinstall.json"
	os.WriteFile(outPath, jsonData, 0644)

	fmt.Printf("\033[1;32m[SUCCESS]\033[0m Flight plan ready at %s\n", outPath)
	ExecutePlan(plan)
}
