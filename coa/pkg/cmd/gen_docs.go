package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docTarget string

// genDocsCmd è un comando nascosto usato dal Makefile/builder per generare la documentazione
var genDocsCmd = &cobra.Command{
	Use:    "_gen_docs",
	Hidden: true, // Nascondiamo questo comando agli utenti normali (non apparirà in 'coa --help')
	Short:  "Genera file Markdown, Man pages e script di autocompletamento",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("[gen_docs] Generazione documentazione nella cartella: %s\n", docTarget)

		// 1. Preparazione delle cartelle di destinazione
		mdDir := filepath.Join(docTarget, "md")
		manDir := filepath.Join(docTarget, "man")
		compDir := filepath.Join(docTarget, "completion")

		dirs := []string{mdDir, manDir, compDir}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("[ERRORE] Impossibile creare la directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		// 2. Generazione Documentazione Markdown
		if err := doc.GenMarkdownTree(rootCmd, mdDir); err != nil {
			fmt.Printf("[ERRORE] Generazione Markdown fallita: %v\n", err)
			os.Exit(1)
		}

		// 3. Generazione Pagine Man (Manuali Linux classici)
		header := &doc.GenManHeader{
			Title:   "COA",
			Section: "1",
		}
		if err := doc.GenManTree(rootCmd, header, manDir); err != nil {
			fmt.Printf("[ERRORE] Generazione Man pages fallita: %v\n", err)
			os.Exit(1)
		}

		// 4. Generazione Script di Autocompletamento (Bash, Zsh, Fish)
		rootCmd.GenBashCompletionFile(filepath.Join(compDir, "coa.bash"))
		rootCmd.GenZshCompletionFile(filepath.Join(compDir, "coa.zsh"))
		rootCmd.GenFishCompletionFile(filepath.Join(compDir, "coa.fish"), true)

		fmt.Println("[gen_docs] ✅ Documentazione e completamenti generati con successo.")
	},
}

func init() {
	genDocsCmd.Flags().StringVar(&docTarget, "target", "./docs", "Target directory per la generazione")
	rootCmd.AddCommand(genDocsCmd)
}
