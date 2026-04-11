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

func GetDiscovery() DiscoveryData {
	d := distro.NewDistro()

	// Nota: in coa, i campi si chiamano Distro e Family (con la maiuscola)
	return DiscoveryData{
		DistroName:   d.Distro, // <--- Cambiato da d.Name a d.Distro
		Family:       d.Family, // Questo dovrebbe essere già corretto se è maiuscolo
		Architecture: runtime.GOARCH,
		IsRoot:       os.Geteuid() == 0,
	}
}
