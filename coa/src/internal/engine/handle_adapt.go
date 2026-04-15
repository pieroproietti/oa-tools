package engine

import (
	"fmt"
	"os/exec"
)

// HandleAdapt adatta la risoluzione del monitor per le Virtual Machine
func HandleAdapt() {
	fmt.Println("\033[1;33m[coa]\033[0m Adapting monitor resolution...")
	virtualOutputs := []string{"Virtual-0", "Virtual-1", "Virtual-2", "Virtual-3"}
	for _, output := range virtualOutputs {
		cmd := exec.Command("xrandr", "--output", output, "--auto")
		_ = cmd.Run() // Ignoriamo gli errori se l'output non esiste
	}
	fmt.Println("\033[1;32m[coa]\033[0m Resolution adapted.")
}
