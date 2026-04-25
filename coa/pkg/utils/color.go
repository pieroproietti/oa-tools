package utils

import "fmt"

const (
	ColorCyan  = "\033[1;36m"
	ColorRed   = "\033[1;31m"
	ColorReset = "\033[0m"
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
