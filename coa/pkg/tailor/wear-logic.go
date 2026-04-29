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

// getAvailablePackages interroga il sistema per ottenere i pacchetti installabili (Debian)
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

// installWithRetries filtra i pacchetti inesistenti prima dell'installazione su Debian
// o genera il prompt AI su altri sistemi.
func installWithRetries(packages []string, retries int) {
	if len(packages) == 0 {
		return
	}

	// 1. Controllo se siamo su Debian/derivata
	if _, err := exec.LookPath("apt-get"); err != nil {
		printAiPrompt(packages)
		return
	}

	// 2. Filtriamo i pacchetti esistenti per evitare che apt blocchi tutto
	available := getAvailablePackages()
	var toInstall []string
	var missing []string

	if available != nil {
		for _, pkg := range packages {
			if _, ok := available[pkg]; ok {
				toInstall = append(toInstall, pkg)
			} else {
				missing = append(missing, pkg)
			}
		}
	} else {
		toInstall = packages // Fallback se apt-cache fallisce
	}

	if len(missing) > 0 {
		logToFile(fmt.Sprintf("ATTENZIONE: %d pacchetti saltati (non trovati): %v", len(missing), missing))
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

// printAiPrompt genera un prompt ricco di informazioni e crea il file AIPrompt.txt nella home
func printAiPrompt(packages []string) {
	d := distro.NewDistro()
	logToFile(fmt.Sprintf("Sistema %s rilevato (Non-Debian). Generazione prompt e file AIPrompt.txt...", d.DistroLike))

	// 1. Cattura info Hardware VGA/3D (fondamentale per driver e KMS)
	gpuCmd := "lspci -k | grep -A 2 -E 'VGA|3D'"
	gpuInfo, _ := exec.Command("sh", "-c", gpuCmd).Output()

	// 2. Cattura sessioni X11 disponibili (per configurazione LightDM)
	sessionCmd := "ls /usr/share/xsessions/ 2>/dev/null"
	sessions, _ := exec.Command("sh", "-c", sessionCmd).Output()

	// 3. Costruzione della stringa del prompt
	var sb strings.Builder
	sb.WriteString("--- PROMPT PER L'ASSISTENTE AI ---\n")
	sb.WriteString(fmt.Sprintf("Sto usando %s (base %s).\n", d.DistroID, d.DistroLike))
	sb.WriteString(fmt.Sprintf("Ho bisogno di installare e configurare questi pacchetti:\n%s\n\n", strings.Join(packages, " ")))

	sb.WriteString("INFO HARDWARE (per driver video e KMS):\n")
	if len(gpuInfo) > 0 {
		sb.WriteString(string(gpuInfo))
	} else {
		sb.WriteString("Nessuna info VGA trovata (pciutils non installato?).\n")
	}

	sb.WriteString("\nSESSIONI DESKTOP DISPONIBILI:\n")
	if len(sessions) > 0 {
		sb.WriteString(string(sessions))
	} else {
		sb.WriteString("Nessuna sessione trovata in /usr/share/xsessions/\n")
	}

	sb.WriteString("\nPer favore, dammi il comando esatto per installare i pacchetti equivalenti su questa distro e i passi necessari per configurare LightDM correttamente.\n")
	sb.WriteString("----------------------------------------\n")

	promptContent := sb.String()

	// 4. Stampa a video per l'utente
	fmt.Println("\n" + utils.ColorCyan + promptContent + utils.ColorReset)

	// 5. Scrittura su file AIPrompt.txt nella HOME reale dell'utente
	userHome, _ := os.UserHomeDir()
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		userHome = filepath.Join("/home", sudoUser)
	}

	promptFile := filepath.Join(userHome, "AIPrompt.txt")
	err := os.WriteFile(promptFile, []byte(promptContent), 0644)

	if err != nil {
		logToFile(fmt.Sprintf("Errore durante la creazione di AIPrompt.txt: %v", err))
	} else {
		// Se siamo sotto sudo, ripristiniamo l'owner del file all'utente reale
		if sudoUser != "" {
			utils.Exec(fmt.Sprintf("sudo chown %s:%s %s", sudoUser, sudoUser, promptFile))
		}
		logToFile(fmt.Sprintf("✅ File AIPrompt.txt generato in: %s", promptFile))
		fmt.Printf("File di prompt generato nella Home: %s%s%s\n\n", utils.ColorYellow, promptFile, utils.ColorReset)
	}
}
