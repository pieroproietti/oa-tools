// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package distro

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

// Distro rappresenta le informazioni dell'ambiente host "purificate"
type Distro struct {
	DistroID       string
	CodenameID     string
	ReleaseID      string
	FamilyID       string
	DistroLike     string
	DistroUniqueID string
}

// DerivativeMapping mappa la struttura del file YAML
type DerivativeMapping struct {
	DistroLike   string   `json:"distroLike"`
	CodenameLike string   `json:"codenameLike"`
	Family       string   `json:"family"`
	Derivatives  []string `json:"derivatives"`
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

// resolveDerivative cerca l'identità originale nel file YAML
func resolveDerivative(distroID string, codenameID string) (bool, *Distro) {
	// Ordine di ricerca: Locale (sviluppo) -> Sistema (produzione)
	paths := []string{
		"conf/derivatives.yaml",
		"/etc/coa/derivatives.yaml",
		"derivatives.yaml",
	}

	var yamlData []byte
	var err error
	for _, p := range paths {
		yamlData, err = os.ReadFile(p)
		if err == nil {
			break // Trovato!
		}
	}

	if err != nil {
		fmt.Println("\033[1;33m[coa]\033[0m Attenzione: file derivatives.yaml non trovato.")
		return false, nil
	}

	var mappings []DerivativeMapping
	if err := yaml.Unmarshal(yamlData, &mappings); err != nil {
		fmt.Printf("\033[1;31m[coa]\033[0m Errore nel parsing di derivatives.yaml: %v\n", err)
		return false, nil
	}

	for _, mapping := range mappings {
		for _, deriv := range mapping.Derivatives {
			if strings.EqualFold(deriv, distroID) || strings.EqualFold(deriv, codenameID) {
				return true, &Distro{
					DistroID:       distroID,
					CodenameID:     codenameID,
					FamilyID:       mapping.Family,
					DistroLike:     mapping.DistroLike,
					DistroUniqueID: mapping.CodenameLike,
				}
			}
		}
	}
	return false, nil
}

// NewDistro inizializza e riconosce la Distro
func NewDistro() *Distro {
	osInfo := parseOsRelease()

	rawID := osInfo["ID"]
	rawCodename := osInfo["VERSION_CODENAME"]
	rawRelease := osInfo["VERSION_ID"]

	idLower := strings.ToLower(rawID)

	d := &Distro{
		DistroID:   rawID,
		CodenameID: rawCodename,
		ReleaseID:  rawRelease,
	}

	switch idLower {
	case "debian":
		d.FamilyID = "debian"
		d.DistroLike = "Debian"
		d.DistroUniqueID = rawCodename
		return d
	case "arch":
		d.FamilyID = "archlinux"
		d.DistroLike = "Arch"
		d.DistroUniqueID = "rolling"
		return d
	case "fedora":
		d.FamilyID = "fedora"
		d.DistroLike = "Fedora"
		d.DistroUniqueID = "rolling"
		return d
	}

	found, derivativeDistro := resolveDerivative(rawID, rawCodename)
	if found {
		return derivativeDistro
	}

	fmt.Printf("\033[1;31m[coa]\033[0m Distro sconosciuta (%s/%s). Aggiungila a derivatives.yaml!\n", rawID, rawCodename)
	os.Exit(1)
	return nil
}
