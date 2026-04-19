package pilot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// --- DEFINIZIONE DELLE STRUTTURE (Le aree che mancavano) ---
type RemasterConfig struct {
	BootParams string            `json:"boot_params"`
	IsoLinks   map[string]string `json:"iso_links,omitempty"`
	AdminGroup string            `json:"admin_group"`
	UserGroups []string          `json:"user_groups"`
}

type InitrdTask struct {
	Command    string
	SetupFiles map[string]string
	Remaster   RemasterConfig
}

// Struttura interna per mappare l'unione dei file YAML
type brainInternal struct {
	Initrd struct {
		Live interface{} `json:"live"`
	} `json:"initrd"`
	Identity struct {
		AdminGroup string   `json:"admin_group"`
		UserGroups []string `json:"user_groups"`
	} `json:"identity"`
	Boot struct {
		Params string `json:"params"`
	} `json:"boot"`
	Layout struct {
		Links map[string]string `json:"links"`
	} `json:"layout"`
}

// --- LOGICA DEL PILOTA ---

func GetInitrdTask(familyID string) *InitrdTask {
	basePath := findBrainDir()
	if basePath == "" {
		return nil
	}

	mappingData, err := os.ReadFile(filepath.Join(basePath, "distro.yaml"))
	if err != nil {
		return nil
	}

	var mapping struct {
		Families map[string]string `yaml:"families"`
	}
	yaml.Unmarshal(mappingData, &mapping)

	folderName := mapping.Families[familyID]
	if folderName == "" {
		folderName = familyID + ".d"
	}

	familyPath := filepath.Join(basePath, folderName)

	var bi brainInternal
	files, _ := filepath.Glob(filepath.Join(familyPath, "*.yaml"))

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		// IMPORTANTE: Controlliamo se lo YAML è valido
		if err := yaml.Unmarshal(data, &bi); err != nil {
			fmt.Printf("[PILOT ERROR] Failed to parse %s: %v\n", file, err)
			continue
		}
	}

	task := &InitrdTask{
		SetupFiles: make(map[string]string),
		Remaster: RemasterConfig{
			BootParams: bi.Boot.Params,
			IsoLinks:   bi.Layout.Links,
			AdminGroup: bi.Identity.AdminGroup,
			UserGroups: bi.Identity.UserGroups,
		},
	}

	parseInitrd(bi.Initrd.Live, task)
	return task
}

// --- FUNZIONI DI SUPPORTO ---

func findBrainDir() string {
	paths := []string{"coa/conf/brain.d", "conf/brain.d", "/etc/coa/brain.d", "brain.d"}
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return ""
}

func parseInitrd(live interface{}, task *InitrdTask) {
	if cmd, ok := live.(string); ok {
		task.Command = cmd
		return
	}
	if m, ok := live.(map[string]interface{}); ok {
		if cmd, ok := m["command"].(string); ok {
			task.Command = cmd
		}
		if files, ok := m["setup_files"].(map[string]interface{}); ok {
			for path, content := range files {
				task.SetupFiles[path] = content.(string)
			}
		}
	}
}

// RunBrainLint scansiona brain.d e aggiunge gli header mancanti
func RunBrainLint() {
	basePath := findBrainDir()
	if basePath == "" {
		fmt.Println("[ERRORE] Cartella brain.d non trovata nei percorsi standard.")
		return
	}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Saltiamo le directory e i file che non sono YAML (o il file delle mappature)
		if info.IsDir() || filepath.Ext(path) != ".yaml" || filepath.Base(path) == "distro.yaml" {
			return nil
		}

		// 1. Estrazione metadati dal percorso
		// relPath sarà qualcosa come "archlinux.d/initrd.yaml"
		relPath, _ := filepath.Rel(basePath, path)
		dirName := filepath.Dir(relPath)
		fileName := filepath.Base(relPath)

		family := strings.TrimSuffix(dirName, ".d")
		area := strings.TrimSuffix(fileName, ".yaml")

		fmt.Printf("[LINT] Controllo unità: %s/%s\n", family, area)

		// 2. Lettura e verifica dell'header
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("impossibile leggere %s: %v", path, err)
		}

		strContent := string(content)

		// Cerchiamo il marcatore unico di coa
		if !strings.Contains(strContent, "[ coa brain unit ]") {
			// Costruzione del "cappello"
			header := "# [ coa brain unit ]\n"
			header += fmt.Sprintf("# family: %s\n", family)
			header += fmt.Sprintf("# area:   %s\n", area)
			header += "# --------------------------------------------------\n\n"

			// Uniamo l'header al contenuto originale (pulendo spazi bianchi extra in cima)
			newContent := header + strings.TrimSpace(strContent) + "\n"

			// 3. Scrittura del file "vestito"
			err = os.WriteFile(path, []byte(newContent), 0644)
			if err != nil {
				fmt.Printf("  └─ [ERRORE] Scrittura fallita: %v\n", err)
			} else {
				fmt.Println("  └─ Cappello aggiunto con successo!")
			}
		} else {
			fmt.Println("  └─ Unità già configurata correttamente.")
		}

		return nil
	})

	if err != nil {
		fmt.Printf("[ERRORE] Durante il lint del cervello: %v\n", err)
	} else {
		fmt.Println("\n[SUCCESSO] Il Cervello è ora ordinato e riconoscibile.")
	}
}

// GenerateBootConfig crea il file grub.cfg basandosi sui dati del brain
func GenerateBootConfig(familyID string, task *InitrdTask) error {
	label := "OA_LIVE" // Potrebbe essere dinamico in futuro
	params := task.Remaster.BootParams

	// percorsi di kernel/initrd
	kernelPath := "/live/vmlinuz"
	initrdPath := "/live/initrd.img"

	// Template GRUB
	content := fmt.Sprintf(`search --no-floppy --set=root --label %s
set default="0"
set timeout=10

menuentry "oa Live System (%s)" {
    echo "Loading kernel..."
    linux %s %s
    echo "Loading initial ramdisk..."
    initrd %s
}
`, label, familyID, kernelPath, params, initrdPath)

	// Assicuriamoci che la cartella temporanea esista
	os.MkdirAll("/tmp/coa", 0755)
	return os.WriteFile("/tmp/coa/grub.cfg.final", []byte(content), 0644)
}
