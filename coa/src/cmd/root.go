package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// AppVersion verrà iniettata in fase di compilazione tramite i ldflags (es. -X 'coa/src/cmd.AppVersion=v0.5.0')
var AppVersion = "dev"

var rootCmd = &cobra.Command{
	Use:   "coa",
	Short: "coa (brooding in my dialect) - The Mind orchestrator",
	Long:  "coa is the orchestrator written in Go, designed to be the \"Mind\" behind the GNU/Linux system remastering process.",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// Execute viene chiamato dal main.go per avviare il parsing della CLI
func Execute() error {
	return rootCmd.Execute()
}

// CheckSudoRequirements è un helper condiviso tra i comandi per verificare i privilegi
func CheckSudoRequirements(cmdName string, shouldBeRoot bool) {
	isRoot := os.Geteuid() == 0

	if shouldBeRoot && !isRoot {
		fmt.Printf("\n\033[1;31m[ERROR]\033[0m The command 'coa %s' requires root privileges.\n", cmdName)
		fmt.Printf("Please use: \033[1msudo coa %s\033[0m\n\n", cmdName)
		os.Exit(1)
	}

	if !shouldBeRoot && isRoot {
		fmt.Printf("\n\033[1;31m[ERROR]\033[0m Do not run 'coa %s' with sudo.\n", cmdName)
		fmt.Printf("Run it as a normal user: \033[1mcoa %s\033[0m\n\n", cmdName)
		os.Exit(1)
	}
}
