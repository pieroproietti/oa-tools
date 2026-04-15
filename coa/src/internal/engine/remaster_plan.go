// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package engine

import (
	"coa/src/internal/distro"
	"fmt"
)

// GeneratePlan costruisce il piano di volo dinamico
func GeneratePlan(d *distro.Distro, mode string, workPath string) FlightPlan {
	plan := FlightPlan{
		PathLiveFs:      workPath,
		Mode:            mode,
		Family:          d.FamilyID,
		BootloadersPath: BootloaderRoot,
	}

	bootParams := "boot=live components"
	switch d.FamilyID {
	case "archlinux":
		bootParams = "archisobasedir=arch archisolabel=OA_LIVE"
	case "fedora", "rhel", "centos", "rocky", "almalinux", "opensuse":
		bootParams = "root=live:CDLABEL=OA_LIVE rd.live.image rd.live.dir=live rd.live.squashimg=filesystem.squashfs selinux=0"
	}

	if mode == "standard" {
		adminGroup := "sudo"
		if d.FamilyID == "archlinux" || d.FamilyID == "fedora" || d.FamilyID == "rhel" || d.FamilyID == "centos" || d.FamilyID == "rocky" || d.FamilyID == "almalinux" {
			adminGroup = "wheel"
		}

		plan.Users = []UserConfig{
			{
				Login:    "live",
				Password: "$6$wM.wY0QtatvbQMHZ$QtIKXSpIsp2Sk57.Ny.JHk7hWDu.lxPtUYaTOiBnP4WBG5KS6JpUlpXj2kcSaaMje7fr01uiGmxZhE8kfZRqv.",
				Gecos:    "live,,,",
				Home:     "/home/live",
				Shell:    "/bin/bash",
				Groups:   []string{"cdrom", "audio", "video", "plugdev", "netdev", "autologin", adminGroup},
			},
		}

		plan.Plan = []Action{
			{Command: "oa_remaster_users"},
		}

		// Comando pulito: OA gestisce il chroot nativamente
		sudoersCmd := fmt.Sprintf("mkdir -p /etc/sudoers.d && echo '%%%s ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/00-oa-live && chmod 0440 /etc/sudoers.d/00-oa-live", adminGroup)

		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_shell",
			RunCommand: sudoersCmd,
			Chroot:     true,
		})

	} else {
		plan.Users = []UserConfig{}
		plan.Plan = []Action{
			{Command: "oa_remaster_users"},
		}
	}

	if d.FamilyID == "fedora" || d.FamilyID == "rhel" || d.FamilyID == "centos" || d.FamilyID == "rocky" || d.FamilyID == "almalinux" {
		targetConfDir := "/etc/dracut.conf.d"
		targetConfPath := fmt.Sprintf("%s/coa.conf", targetConfDir)
		dracutConfig := "hostonly=\"no\"\nadd_dracutmodules+=\" dmsquash-live rootfs-block bash \"\ncompress=\"xz\""

		writeCmd := fmt.Sprintf("mkdir -p %s && echo -e '%s' > %s", targetConfDir, dracutConfig, targetConfPath)

		plan.Plan = append(plan.Plan, Action{
			Command:    "oa_sys_shell",
			RunCommand: writeCmd,
			Chroot:     true,
		})
	}

	// Grazie al chroot nativo in OA, le variabili vengono risolte correttamente nel guest
	// Comando specifico per la rigenerazione dell'initrd via shell.
	// Ora che /boot è una copia fisica e /tmp è un tmpfs (gestiti da oa),
	// il comando torna a essere pulito e lineare.
	var shellInitrdCmd string
	switch d.FamilyID {
	case "archlinux":
		// Grazie alla copia fisica in C, /boot è ora una directory reale e scrivibile.
		// Ci limitiamo a garantire i permessi corretti per sicurezza.
		shellInitrdCmd = "chmod 755 /boot && " +
			"KVER=$(ls /lib/modules | head -n 1) && " +
			"mkinitcpio -c /etc/coa_mkinitcpio.conf -k $KVER -g /boot/initrd.img"

	case "fedora", "rhel", "centos", "rocky", "almalinux":
		shellInitrdCmd = "dracut --force --regenerate-all"

	default: // Debian/Ubuntu
		// update-initramfs gestisce correttamente i vari kernel installati
		shellInitrdCmd = "update-initramfs -u -k all || update-initramfs -c -k all"
	}

	excludeFilePath := generateExcludeList(mode)

	plan.Plan = append(plan.Plan,
		Action{
			Command:    "oa_sys_shell",
			RunCommand: shellInitrdCmd,
			Chroot:     true,
		},
		Action{Command: "oa_remaster_livestruct"},
		Action{Command: "oa_remaster_isolinux", BootParams: bootParams},
		Action{Command: "oa_remaster_uefi", BootParams: bootParams},
		Action{
			Command:     "oa_remaster_squash",
			ExcludeList: excludeFilePath,
		},
	)

	if mode == "crypted" {
		plan.Plan = append(plan.Plan, Action{
			Command:         "oa_remaster_crypted",
			CryptedPassword: "evolution",
		})
	}

	// Chiamata a getIsoName per ottenere il nome della ISO
	isoName := getIsoName(d)

	plan.Plan = append(plan.Plan, Action{
		Command:   "oa_remaster_iso",
		VolID:     "OA_LIVE",
		OutputISO: isoName,
	})

	plan.Plan = append(plan.Plan, Action{Command: "oa_remaster_cleanup"})

	return plan
}
