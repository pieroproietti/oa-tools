package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"coa/pkg/distro"
	"coa/pkg/engine"
	"coa/pkg/pilot"
	"coa/pkg/utils"

	"github.com/spf13/cobra"
)

// --- SISTEMA DI LOGGING CENTRALIZZATO ---
const (
	ColorCyan   = "\033[1;36m"
	ColorRed    = "\033[1;31m"
	ColorGreen  = "\033[1;32m"
	ColorYellow = "\033[1;33m" // Aggiunto per i messaggi di debug/breakpoint
	ColorReset  = "\033[0m"
)

func LogCoala(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[coa]%s %s\n", ColorCyan, ColorReset, msg)
}

func LogSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s[coa]%s %s\n", ColorGreen, ColorReset, msg)
}

func LogError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\n%s[ERRORE]%s %s\n", ColorRed, ColorReset, msg)
}

// ----------------------------------------

var (
	produceMode string
	producePath string
	stopAfter   string // NUOVA VARIABILE: Memorizza il target del breakpoint
)

var remasterCmd = &cobra.Command{
	Use:   "remaster",
	Short: "Start a system remastering flight (ISO production)",
	Long: `The 'remaster' command orchestrates the creation of a bootable live ISO. 
It uses the new Coala architecture to read the agnostic Brain profile 
and generate a precise execution plan for the OA engine.`,
	Example: `  # Start a standard ISO remastering
  sudo ./coa remaster --mode standard
  
  # Debug mode: stop after a specific step
  sudo ./coa remaster --stop-after coa-initrd`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)

		LogCoala("Avvio procedura di rimasterizzazione...")

		// 1. Identità: Chi siamo?
		myDistro := distro.NewDistro()
		isoName := myDistro.GetISOName()
		finalPath := filepath.Join(producePath, isoName)
		LogCoala("L'uovo verrà generato in: %s", finalPath)

		// 2. PILOT: Carichiamo lo spartito dal Brain
		profile, err := pilot.DetectAndLoad()
		if err != nil {
			LogError("Impossibile caricare il Brain Profile: %v", err)
			os.Exit(1)
		}
		LogSuccess("Spartito caricato con successo.")

		// 3. ENGINE: Generiamo il piano JSON per oa
		// NOTA: Passiamo stopAfter come 5° parametro al motore!
		planPath, err := engine.GeneratePlan(profile.Remaster, myDistro.FamilyID, true, producePath, finalPath, stopAfter)
		// planPath, err := engine.GeneratePlan(profile.Remaster, myDistro.FamilyID, true, producePath, stopAfter)
		if err != nil {
			LogError("Impossibile generare il piano di volo: %v", err)
			os.Exit(1)
		}

		// RECUPERO BOOTLOADERS
		LogCoala("Recupero bootloaders (penguins-bootloaders)...")
		utils.EnsureBootloaders("/tmp/coa/bootloaders")

		// GENERAZIONE EXCLUSIONI
		LogCoala("Generazione lista di esclusione (%s mode)...", produceMode)
		engine.GenerateExcludeList(produceMode)

		// 4. DECOLLO: Eseguiamo il motore C (oa) passandogli il JSON appena generato
		LogCoala("Passaggio dei comandi al motore OA...")
		oaCmd := exec.Command("oa", planPath)

		// Colleghiamo l'output di oa direttamente al terminale dell'utente
		oaCmd.Stdout = os.Stdout
		oaCmd.Stderr = os.Stderr

		if err := oaCmd.Run(); err != nil {
			LogError("L'esecuzione di oa è fallita: %v", err)
			os.Exit(1)
		}

		// Vittoria finale: Differenziamo il messaggio se abbiamo usato il breakpoint
		if stopAfter != "" {
			fmt.Printf("\n%s[DEBUG]%s Breakpoint raggiunto e ambiente smontato in sicurezza. Pronto per l'ispezione! 🐧🔍\n", ColorYellow, ColorReset)
		} else {
			fmt.Printf("\n%s[SUCCESSO]%s Rimasterizzazione completata! L'uovo è pronto. 🐧🥚\n", ColorGreen, ColorReset)
		}
	},
}

func init() {
	remasterCmd.Flags().StringVar(&produceMode, "mode", "standard", "standard, clone, or crypted")
	remasterCmd.Flags().StringVar(&producePath, "path", "/home/eggs", "working directory")

	// Registrazione del nuovo flag per il breakpoint
	remasterCmd.Flags().StringVar(&stopAfter, "stop-after", "", "Ferma l'esecuzione dopo uno step specifico (es. coa-initrd)")

	rootCmd.AddCommand(remasterCmd)
}
