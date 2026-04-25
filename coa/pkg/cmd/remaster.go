package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"coa/pkg/distro"
	"coa/pkg/engine"
	"coa/pkg/pilot"
	"coa/pkg/utils"

	"github.com/spf13/cobra"
)

// --- SISTEMA DI LOGGING CENTRALIZZATO ---
const (
	ColorCyan  = "\033[1;36m"
	ColorRed   = "\033[1;31m"
	ColorGreen = "\033[1;32m"
	ColorReset = "\033[0m"
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
)

var remasterCmd = &cobra.Command{
	Use:   "remaster",
	Short: "Start a system remastering flight (ISO production)",
	Long: `The 'remaster' command orchestrates the creation of a bootable live ISO. 
It uses the new Coala architecture to read the agnostic Brain profile 
and generate a precise execution plan for the OA engine.`,
	Example: `  # Start a standard ISO remastering
  sudo ./coa remaster --mode standard`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)

		LogCoala("Avvio procedura di rimasterizzazione...")

		// 1. Identità: Chi siamo?
		myDistro := distro.NewDistro()

		// 2. PILOT: Carichiamo lo spartito dal Brain
		profile, err := pilot.DetectAndLoad()
		if err != nil {
			LogError("Impossibile caricare il Brain Profile: %v", err)
			os.Exit(1)
		}
		LogSuccess("Spartito caricato con successo.")

		// 3. ENGINE: Generiamo il piano JSON per oa
		err = engine.GeneratePlan(profile.Remaster, myDistro.FamilyID, true, producePath)
		if err != nil {
			LogError("Impossibile generare il piano di volo: %v", err)
			os.Exit(1)
		}

		// --- RECUPERO BOOTLOADERS ---
		LogCoala("Recupero bootloaders (penguins-bootloaders)...")
		utils.EnsureBootloaders("/tmp/coa/bootloaders")

		// 4. DECOLLO: Eseguiamo il motore C (oa) passandogli il JSON appena generato
		LogCoala("Passaggio dei comandi al motore OA...")
		oaCmd := exec.Command("oa", "oa-plan.json")

		// Colleghiamo l'output di oa direttamente al terminale dell'utente
		oaCmd.Stdout = os.Stdout
		oaCmd.Stderr = os.Stderr

		if err := oaCmd.Run(); err != nil {
			LogError("L'esecuzione di oa è fallita: %v", err)
			os.Exit(1)
		}

		// Vittoria finale
		fmt.Printf("\n%s[SUCCESSO]%s Rimasterizzazione completata! L'uovo è pronto. 🐧🥚\n", ColorGreen, ColorReset)
	},
}

func init() {
	remasterCmd.Flags().StringVar(&produceMode, "mode", "standard", "standard, clone, or crypted")
	remasterCmd.Flags().StringVar(&producePath, "path", "/home/eggs", "working directory")

	rootCmd.AddCommand(remasterCmd)
}
