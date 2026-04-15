package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// HandleDocs genera in automatico la wiki markdown, le man pages e gli script di autocompletamento.
func HandleDocs(rootCmd *cobra.Command, targetDir string) error {
	fmt.Printf("\033[1;34m[coa docs]\033[0m Generating project documentation in %s...\n", targetDir)

	// 1. Definizione e creazione dei percorsi
	mdPath := filepath.Join(targetDir, "md")
	manPath := filepath.Join(targetDir, "man")
	compPath := filepath.Join(targetDir, "completion")

	if err := os.MkdirAll(mdPath, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(manPath, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(compPath, 0755); err != nil {
		return err
	}

	// 2. Generazione Markdown
	if err := doc.GenMarkdownTree(rootCmd, mdPath); err != nil {
		return fmt.Errorf("markdown generation failed: %w", err)
	}
	fmt.Printf("\033[1;32m[+] Markdown Wiki generated in %s\033[0m\n", mdPath)

	// 3. Generazione Man Pages
	header := &doc.GenManHeader{
		Title:   "COA",
		Section: "1",
		Source:  "penguins-eggs",
		Manual:  "coa - The Artisan Orchestrator",
	}
	if err := doc.GenManTree(rootCmd, header, manPath); err != nil {
		return fmt.Errorf("man pages generation failed: %w", err)
	}
	fmt.Printf("\033[1;32m[+] Man pages generated in %s\033[0m\n", manPath)

	// 4. Generazione Autocompletion
	rootCmd.GenBashCompletionFile(filepath.Join(compPath, "coa.bash"))
	rootCmd.GenZshCompletionFile(filepath.Join(compPath, "coa.zsh"))
	rootCmd.GenFishCompletionFile(filepath.Join(compPath, "coa.fish"), true)

	fmt.Printf("\033[1;32m[+] Autocompletion scripts generated in %s\033[0m\n", compPath)

	// Generiamo l'indice personalizzato
	if err := generateCommandIndex(rootCmd, mdPath); err != nil {
		return fmt.Errorf("failed to generate README index: %w", err)
	}

	fmt.Printf("\033[1;32m[+] README index generated in %s\033[0m\n", mdPath)
	return nil
}

// generateCommandIndex crea un file README.md che funge da indice per la wiki
func generateCommandIndex(rootCmd *cobra.Command, outputDir string) error {
	var builder strings.Builder

	builder.WriteString("# 🛠️ coa Command Reference\n\n")
	builder.WriteString("Benvenuto nella guida ai comandi di **coa**. Qui trovi l'elenco completo delle funzionalità orchestrate dal motore.\n\n")

	// Definiamo delle categorie per organizzare meglio il README
	categories := map[string][]*cobra.Command{
		"🚀 Core Actions":  {},
		"📦 Export & Sync": {},
		"⚙️ System Tools": {},
	}

	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden {
			continue
		}

		// Logica di smistamento nelle categorie
		switch cmd.Name() {
		case "remaster", "sysinstall": // Aggiornati i nomi qui
			categories["🚀 Core Actions"] = append(categories["🚀 Core Actions"], cmd)
		case "export":
			categories["📦 Export & Sync"] = append(categories["📦 Export & Sync"], cmd)
		default:
			categories["⚙️ System Tools"] = append(categories["⚙️ System Tools"], cmd)
		}
	}

	// Scriviamo le categorie nel file
	for catName, cmds := range categories {
		if len(cmds) == 0 {
			continue
		}
		builder.WriteString(fmt.Sprintf("## %s\n\n", catName))
		for _, cmd := range cmds {
			// Il link punta al file generato da Cobra (es: coa_produce.md)
			link := fmt.Sprintf("coa_%s.md", cmd.Name())
			builder.WriteString(fmt.Sprintf("- [**%s**](%s) - %s\n", cmd.Name(), link, cmd.Short))
		}
		builder.WriteString("\n")
	}

	// Salviamo il file README.md nella cartella dei Markdown
	indexPath := filepath.Join(outputDir, "README.md")
	return os.WriteFile(indexPath, []byte(builder.String()), 0644)
}
