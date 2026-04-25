package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
)

// IsUEFI controlla se il sistema è avviato in modalità UEFI
func IsUEFI() bool {
	_, err := os.Stat("/sys/firmware/efi")
	return !os.IsNotExist(err)
}

// GetDistroID legge /etc/os-release per determinare la famiglia della distribuzione
func GetDistroID() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "debian" // Default di emergenza
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			// Estrae il valore rimuovendo "ID=" e le eventuali virgolette
			id := strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, "\"")
			return id
		}
	}
	return "debian"
}

// GenerateFinalizePlan crea il JSON per l'ultimo step di Calamares
func GenerateFinalizePlan() error {
	isUEFI := IsUEFI()
	distro := GetDistroID()

	// 1. Configurazione comando GRUB
	var grubCmd string
	if isUEFI {
		grubCmd = "grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=OA > /var/log/grub-debug.log 2>&1 && grub-mkconfig -o /boot/grub/grub.cfg >> /var/log/grub-debug.log 2>&1"
	} else {
		grubCmd = "grub-install /dev/sda && grub-mkconfig -o /boot/grub/grub.cfg >> /var/log/grub-debug.log 2>&1"
	}

	// 2. Configurazione comando Initramfs dinamico
	var initramfsCmd string
	switch distro {
	case "arch", "manjaro", "endeavouros":
		initramfsCmd = "mkinitcpio -P"
	case "fedora":
		initramfsCmd = "dracut --force --no-hostonly"
	default: // debian, ubuntu, linuxmint, ecc.
		initramfsCmd = "update-initramfs -u -k all"
	}

	// 3. Creazione del FlightPlan
	plan := FlightPlan{
		Mode:       "install",
		PathLiveFs: "/tmp/coa/calamares-root",
		Plan: []Action{
			// Azione 1: Generazione Initramfs (DEVE precedere GRUB)
			{
				Command:    "oa_shell",
				Info:       "Generazione initramfs (" + distro + ")",
				RunCommand: initramfsCmd,
				Chroot:     true,
			},
			// Azione 2: Installazione GRUB
			{
				Command:    "oa_shell",
				Info:       "Installazione bootloader (GRUB)",
				RunCommand: grubCmd,
				Chroot:     true,
			},
		},
	}

	// 4. Scrittura del JSON
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(plan)
	if err != nil {
		return err
	}

	os.MkdirAll("/tmp/coa", 0755)
	return os.WriteFile("/tmp/coa/finalize-plan.json", buf.Bytes(), 0644)
}
