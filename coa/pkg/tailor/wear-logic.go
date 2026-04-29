package tailor

import (
	"bufio"
	"coa/pkg/distro"
	"coa/pkg/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// logToFile scrive un messaggio sia sul log di sistema che su un file locale
func logToFile(message string) {
	utils.LogCoala(message)
	
	logPath := "/var/log/coa-tailor.log"
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Se non possiamo scrivere sul log, procediamo comunque
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

// findYaml cerca esclusivamente il file index.yaml
func findYaml(costumePath string) string {
	fullPath := filepath.Join(costumePath, "index.yaml")
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	}
	return ""
}

// loadSuit trasforma il file YAML fisico nella struttura Suit
func loadSuit(yamlFile string) (*Suit, error) {
	if yamlFile == "" {
		return nil, fmt.Errorf("file 'index.yaml' non trovato")
	}

	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}

	var suit Suit
	if err := yaml.Unmarshal(data, &suit); err != nil {
		return nil, err
	}

	return &suit, nil
}

// getAvailablePackages interroga il sistema per ottenere i pacchetti installabili
func getAvailablePackages() map[string]struct{} {
	available := make(map[string]struct{})

	if _, err := exec.LookPath("apt-cache"); err != nil {
		return nil
	}

	logToFile("Aggiornamento database pacchetti disponibili...")
	cmd := exec.Command("/usr/bin/apt-cache", "pkgnames")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return available
	}

	if err := cmd.Start(); err != nil {
		return available
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			available[line] = struct{}{}
		}
	}
	cmd.Wait()
	return available
}

// installWithRetries filtra i pacchetti inesistenti prima dell'installazione
func installWithRetries(packages []string, retries int) {
	if len(packages) == 0 {
		return
	}

	// 1. Controllo se siamo su Debian/derivata
	if _, err := exec.LookPath("apt-get"); err != nil {
		printAiPrompt(packages)
		return
	}

	// 2. Filtriamo i pacchetti esistenti
	available := getAvailablePackages()
	var toInstall []string
	var missing []string

	for _, pkg := range packages {
		if _, ok := available[pkg]; ok {
			toInstall = append(toInstall, pkg)
		} else {
			missing = append(missing, pkg)
		}
	}

	if len(missing) > 0 {
		logToFile(fmt.Sprintf("ATTENZIONE: %d pacchetti non trovati nei repository e verranno saltati: %v", len(missing), missing))
	}

	if len(toInstall) == 0 {
		logToFile("Nessun pacchetto valido da installare.")
		return
	}

	// 3. Procediamo con l'installazione dei soli pacchetti validi
	pkgString := strings.Join(toInstall, " ")
	cmd := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", pkgString)

	for i := 1; i <= retries; i++ {
		logToFile(fmt.Sprintf("Tentativo installazione %d di %d...", i, retries))
		if err := utils.Exec(cmd); err == nil {
			logToFile("✅ Installazione pacchetti completata.")
			return
		}
		time.Sleep(2 * time.Second)
	}
	
	logToFile("❌ Errore critico durante l'installazione pacchetti dopo i tentativi.")
}

func printAiPrompt(packages []string) {
	d := distro.NewDistro()
	logToFile(fmt.Sprintf("Sistema %s rilevato (Non-Debian). Generazione prompt AI...", d.DistroLike))
	fmt.Println("\n" + utils.ColorCyan + "--- PROMPT PER L'ASSISTENTE AI ---" + utils.ColorReset)
	fmt.Printf("Sto usando %s. Dammi il comando per installare questi pacchetti:\n%s\n", d.DistroLike, strings.Join(packages, " "))
	fmt.Println(utils.ColorCyan + "----------------------------------" + utils.ColorReset + "\n")
}
