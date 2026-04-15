package pilot

import (
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

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

// brainInternal è la struttura che riflette l'unione di tutti i file .yaml
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

func GetInitrdTask(familyID string) *InitrdTask {
	// Cerchiamo la cartella brain.d partendo dai soliti percorsi
	basePath := findBrainDir()
	if basePath == "" {
		return nil
	}

	// 1. Leggiamo la mappatura delle famiglie
	mappingData, _ := os.ReadFile(filepath.Join(basePath, "distro.yaml"))
	var mapping struct {
		Families map[string]string `yaml:"families"`
	}
	yaml.Unmarshal(mappingData, &mapping)

	folderName := mapping.Families[familyID]
	if folderName == "" {
		folderName = familyID + ".d"
	}

	familyPath := filepath.Join(basePath, folderName)

	// 2. Carichiamo e fondiamo tutti i file .yaml della cartella
	var bi brainInternal
	files, _ := filepath.Glob(filepath.Join(familyPath, "*.yaml"))
	for _, file := range files {
		data, _ := os.ReadFile(file)
		yaml.Unmarshal(data, &bi)
	}

	// 3. Prepariamo il task finale
	task := &InitrdTask{
		SetupFiles: make(map[string]string),
		Remaster: RemasterConfig{
			BootParams: bi.Boot.Params,
			IsoLinks:   bi.Layout.Links,
			AdminGroup: bi.Identity.AdminGroup,
			UserGroups: bi.Identity.UserGroups,
		},
	}

	// Gestione flessibile per l'initrd (stringa o mappa)
	parseInitrd(bi.Initrd.Live, task)

	return task
}

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
