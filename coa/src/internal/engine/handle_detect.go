package engine

import (
	"coa/src/internal/distro"
	"fmt"
)

// HandleDetect mostra le informazioni rilevate sul sistema host
func HandleDetect(d *distro.Distro) {
	fmt.Println("\033[1;34m--- coa distro detect ---\033[0m")
	fmt.Printf("Host Distro:     %s\n", d.DistroID)
	fmt.Printf("Family:          %s\n", d.FamilyID)
	fmt.Printf("DistroLike:      %s\n", d.DistroLike)
	fmt.Printf("Codename:        %s\n", d.CodenameID)
	fmt.Printf("Release:         %s\n", d.ReleaseID)
	fmt.Printf("DistroUniqueID:  %s\n", d.DistroUniqueID)
}
