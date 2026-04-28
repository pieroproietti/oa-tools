package cmd

import (
	"github.com/spf13/cobra"
)

// toolsCmd rappresenta il comando padre "tools"
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Strumenti di utilità per la manutenzione e il sistema",
	Long: `Una suite di strumenti ausiliari forniti da coa per la gestione, 
la pulizia e l'ispezione del sistema host e delle ISO.`,
	// Non definiamo una funzione Run, così se l'utente digita solo "coa tools",
	// Cobra stamperà in automatico l'help con la lista dei sotto-comandi (es. clean).
}

func init() {
	// Aggiungiamo tools al comando principale (root)
	// Assicurati che rootCmd sia il nome della variabile del tuo comando principale
	// definito di solito in root.go o main.go
	rootCmd.AddCommand(toolsCmd)
}
