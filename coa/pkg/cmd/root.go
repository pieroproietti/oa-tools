package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// AppVersion viene iniettata dal Makefile durante il build
var AppVersion = "development"

// rootCmd rappresenta il comando base quando viene chiamato senza argomenti
var rootCmd = &cobra.Command{
	Use:   "coa",
	Short: "coa - Calamares & OA Lightweight Architect",
	Long: `coa è l'orchestratore universale per penguins-eggs.
Progettato per essere leggero ed elegante, legge la logica specifica 
delle distribuzioni dal Brain e pilota il motore OA per 
la rimasterizzazione e l'installazione del sistema.`,
	Version: AppVersion,
}

// Execute aggiunge tutti i comandi figli al comando radice e imposta i flag correttamente.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31m[ERRORE]\033[0m %v\n", err)
		os.Exit(1)
	}
}

// CheckSudoRequirements verifica che l'utente sia root se il comando lo richiede
func CheckSudoRequirements(cmdName string, needSudo bool) {
	if needSudo && os.Geteuid() != 0 {
		fmt.Printf("\033[1;33m[AVVISO]\033[0m Il comando '%s' richiede privilegi di root.\n", cmdName)
		fmt.Println("Esegui: sudo coa", cmdName)
		os.Exit(1)
	}
}
