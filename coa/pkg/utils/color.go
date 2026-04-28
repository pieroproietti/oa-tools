package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)


// Utilizziamo lo stesso set di colori
const (
	ColorBlue   = "\033[1;34m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[1;32m"
	ColorRed    = "\033[1;31m"
	ColorReset  = "\033[0m"
	ColorYellow = "\033[33m"
)

// LogCoala stampa un messaggio informativo con il tag [coa] in ciano
func LogCoala(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[coa]%s %s\n", ColorCyan, ColorReset, msg)
}

// LogError stampa un messaggio di errore con il tag [ERRORE] in rosso
func LogError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[ERRORE]%s %s\n", ColorRed, ColorReset, msg)
}

// Usate da tailor
// Exec esegue un comando sh e mostra l'output in tempo reale sul terminale
func Exec(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecQuiet esegue un comando senza mostrare nulla (utile per update veloci)
func ExecQuiet(command string) error {
	cmd := exec.Command("sh", "-c", command)
	return cmd.Run()
}

// ExecCapture esegue un comando e restituisce l'output come stringa
// Fondamentale per getAvailablePackages (apt-cache pkgnames)
func ExecCapture(command string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &out
	// Stderr lo ignoriamo o lo mandiamo a null per pulizia
	return out.String(), cmd.Run()
}
