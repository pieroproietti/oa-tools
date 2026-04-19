package middleware

import (
	"coa/src/internal/distro"
	"os"
	"runtime"
)

type DiscoveryData struct {
	DistroName   string
	Family       string
	Architecture string
	IsRoot       bool
}

// src/middleware/hub.go
func GetDiscovery() DiscoveryData {
	d := distro.NewDistro()

	return DiscoveryData{
		// Cambia d.Distro in d.DistroID
		DistroName: d.DistroID,
		// Cambia d.Family in d.FamilyID
		Family:       d.FamilyID,
		Architecture: runtime.GOARCH,
		IsRoot:       os.Geteuid() == 0,
	}
}
