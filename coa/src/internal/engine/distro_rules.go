package engine

import (
	"coa/src/internal/distro"
	"strings"
)

// Op definisce il tipo di operazione richiesta
type Op string

const (
	OpInitrd       Op = "initrd"
	OpSudoers      Op = "sudoers"
	OpBootParams   Op = "bootparams"
	OpDracutConfig Op = "dracut_conf"
	OpAdminGroup   Op = "admin_group"
)

// Registry contiene la mappatura: FamilyID -> Operazione -> Template/Comando
var rulesRegistry = map[string]map[Op]string{
	"debian": {
		OpInitrd:     "update-initramfs -u -k all || update-initramfs -c -k all",
		OpSudoers:    "mkdir -p /etc/sudoers.d && echo '%sudo ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/00-oa-live && chmod 0440 /etc/sudoers.d/00-oa-live",
		OpBootParams: "boot=live components",
		OpAdminGroup: "sudo",
	},
	"archlinux": {
		OpInitrd:     "chmod 755 /boot && KVER=$(ls /lib/modules | head -n 1) && mkinitcpio -c /etc/coa_mkinitcpio.conf -k $KVER -g /boot/initrd.img",
		OpSudoers:    "mkdir -p /etc/sudoers.d && echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/00-oa-live && chmod 0440 /etc/sudoers.d/00-oa-live",
		OpBootParams: "archisobasedir=arch archisolabel=OA_LIVE",
		OpAdminGroup: "wheel",
	},
	"fedora": {
		OpInitrd:       "dracut --force --regenerate-all",
		OpSudoers:      "mkdir -p /etc/sudoers.d && echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/00-oa-live && chmod 0440 /etc/sudoers.d/00-oa-live",
		OpBootParams:   "root=live:CDLABEL=OA_LIVE rd.live.image rd.live.dir=live rd.live.squashimg=filesystem.squashfs selinux=0",
		OpDracutConfig: "mkdir -p /etc/dracut.conf.d && echo -e 'hostonly=\"no\"\\nadd_dracutmodules+=\" dmsquash-live rootfs-block bash \"\\ncompress=\"xz\"' > /etc/dracut.conf.d/coa.conf",
		OpAdminGroup:   "wheel",
	},
}

// GetRule è la funzione generale. Riceve la distro e l'operazione.
// Supporta anche segnaposti dinamici come {DISK} se necessario.
func GetRule(d *distro.Distro, op Op, vars map[string]string) string {
	familyRules, ok := rulesRegistry[d.FamilyID]
	if !ok {
		// Fallback su debian se la famiglia è sconosciuta
		familyRules = rulesRegistry["debian"]
	}

	command, ok := familyRules[op]
	if !ok {
		return ""
	}

	// Sostituzione dinamica delle variabili (es. per Krill che serve il disco)
	for key, val := range vars {
		placeholder := "{" + strings.ToUpper(key) + "}"
		command = strings.ReplaceAll(command, placeholder, val)
	}

	return command
}
