// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.

package bleach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"coa/pkg/distro" // Assicurati che l'import corrisponda al tuo modulo
)

type Bleach struct {
	Verbose bool
}

func New(verbose bool) *Bleach {
	return &Bleach{Verbose: verbose}
}

func (b *Bleach) log(msg string) {
	if b.Verbose {
		fmt.Printf("\033[1;33m[bleach]\033[0m %s\n", msg)
	}
}

func (b *Bleach) run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if b.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Run()
}

// Clean esegue la pulizia del sistema
func (b *Bleach) Clean() error {
	d := distro.NewDistro()

	b.log(fmt.Sprintf("Inizio pulizia per famiglia: %s", d.FamilyID))

	// 1. Pulizia Gestore Pacchetti in base alla distro
	switch d.FamilyID {
	case "alpine":
		b.run("apk", "cache", "clean")
		b.run("apk", "cache", "purge")

	case "archlinux":
		// Usiamo sh -c per la pipe "yes |"
		b.run("sh", "-c", "yes | pacman -Scc")

	case "debian":
		b.run("apt-get", "clean")
		b.run("apt-get", "autoclean", "-y")
		os.RemoveAll("/var/lib/apt/lists/lock")

	case "fedora", "openmamba":
		b.run("sh", "-c", "dnf remove $(dnf repoquery --installonly --latest-limit=-1 -q) -y")
		b.run("dnf", "clean", "all")

	case "opensuse":
		b.run("zypper", "clean")

	case "voidlinux":
		b.run("xbps-remove", "-O", "-y")
	}

	// 2. Fastpack (Flatpak cache)
	b.log("Pulizia cache Flatpak")
	matches, _ := filepath.Glob("/var/tmp/flatpak-cache-*")
	for _, match := range matches {
		os.RemoveAll(match)
	}

	// 3. Bash History
	b.log("Pulizia cronologia bash")
	os.RemoveAll("/root/.bash_history")

	// 4. Pulizia Journald / Syslog
	b.log("Pulizia log di sistema")
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		b.run("journalctl", "--rotate")
		b.run("journalctl", "--vacuum-time=1s")
	} else {
		// Per i sistemi sysvinit / openrc
		b.run("sh", "-c", "find /var/log -name '*gz' -print0 | xargs -0r rm -f")
		b.run("sh", "-c", "find /var/log/ -type f -exec truncate -s 0 {} \\;")
	}

	// 5. System Cache (PageCache, dentries, inodes)
	b.log("Svuotamento cache Kernel (PageCache, dentries e inodes)")
	b.run("sync")
	os.WriteFile("/proc/sys/vm/drop_caches", []byte("3"), 0644)

	return nil
}
