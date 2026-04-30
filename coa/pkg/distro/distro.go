// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package distro

import (
	"bufio"
	"coa/pkg/utils"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Distro rappresenta le informazioni dell'ambiente host "purificate"
type Distro struct {
	DistroID   string
	CodenameID string
	ReleaseID  string
	FamilyID   string
	DistroLike string
	Arch       string
}

// parseOsRelease legge /etc/os-release
func parseOsRelease() map[string]string {
	info := make(map[string]string)
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return info
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			val := strings.Trim(parts[1], `"'`)
			info[key] = val
		}
	}
	return info
}

// NewDistro inizializza e riconosce la Distro in modo agnostico
func NewDistro() *Distro {
	osInfo := parseOsRelease()

	rawID := strings.ToLower(osInfo["ID"])
	rawLike := strings.ToLower(osInfo["ID_LIKE"])
	likes := strings.Fields(rawLike)

	d := &Distro{
		DistroID:   osInfo["ID"],
		CodenameID: osInfo["VERSION_CODENAME"],
		ReleaseID:  osInfo["VERSION_ID"],
	}

	// Candidati per il riconoscimento: prima l'ID specifico, poi i LIKE
	candidates := append([]string{rawID}, likes...)

	for _, c := range candidates {
		switch c {
		case "debian", "ubuntu", "linuxmint", "kali", "pop":
			d.FamilyID = "debian"
			d.DistroLike = "Debian"
			return d

		case "arch", "endeavouros", "garuda":
			d.FamilyID = "archlinux"
			d.DistroLike = "Arch"
			return d

		case "manjaro", "biglinux", "bigcommunity":
			d.FamilyID = "manjaro"
			d.DistroLike = "Manjaro"
			return d

		case "fedora", "nobara", "rhel", "centos":
			d.FamilyID = "fedora"
			d.DistroLike = "Fedora"
			return d
		}
	}

	// Fallback agnostico
	utils.LogCoala("[coa] Distro non mappata specificamente (%s). Modalità Generica.", osInfo["ID"])
	d.FamilyID = "generic"
	d.DistroLike = osInfo["ID"]
	return d
}

// GetISOName genera il nome completo: egg-of-<distro>-<host>-<arch>-<data>.iso
func (d *Distro) GetISOName() string {
	timestamp := time.Now().Format("2006-01-02_1504")
	prefix := d.GetISOPrefix()
	return fmt.Sprintf("%s%s.iso", prefix, timestamp)
}

// GetISOPrefix genera la parte iniziale: egg-of-<distro>-<host>-<arch>-
func (d *Distro) GetISOPrefix() string {
	distroName := strings.ToLower(strings.ReplaceAll(d.DistroID, " ", "-"))

	codeName := strings.ToLower(strings.ReplaceAll(d.CodenameID, " ", "-"))
	if codeName == "" {
		codeName = strings.ToLower(strings.ReplaceAll(d.ReleaseID, " ", "-"))
	}

	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}
	hostName = strings.ToLower(strings.ReplaceAll(hostName, " ", "-"))

	arch := d.Arch
	if arch == "" {
		arch = runtime.GOARCH
	}

	if codeName == "" {
		return fmt.Sprintf("egg-of-%s-%s-%s-", distroName, hostName, arch)
	}

	return fmt.Sprintf("egg-of-%s-%s-%s-%s-", distroName, codeName, hostName, arch)
}
