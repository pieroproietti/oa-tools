package pilot

import (
	"os"

	"sigs.k8s.io/yaml"
)

type InitrdTask struct {
	Command    string
	SetupFiles map[string]string
}

func findBrainPath() string {
	paths := []string{
		"coa/conf/brain.yaml", // Sviluppo dalla root della repo
		"conf/brain.yaml",     // Esecuzione dalla cartella coa/
		"/etc/coa/brain.yaml", // Produzione (sistema installato)
		"brain.yaml",          // Fallback nella directory corrente
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func GetInitrdTask(family string) *InitrdTask {
	data, err := os.ReadFile(findBrainPath())
	if err != nil {
		return nil
	}

	var brain struct {
		Families map[string]struct {
			Initrd struct {
				Live interface{} `json:"live"`
			} `json:"initrd"`
		} `json:"families"`
	}

	if err := yaml.Unmarshal(data, &brain); err != nil {
		return nil
	}

	f, ok := brain.Families[family]
	if !ok {
		return nil
	}

	task := &InitrdTask{SetupFiles: make(map[string]string)}

	// Gestione flessibile: comando stringa o mappa complessa
	if cmd, ok := f.Initrd.Live.(string); ok {
		task.Command = cmd
		return task
	}

	if m, ok := f.Initrd.Live.(map[string]interface{}); ok {
		if cmd, ok := m["command"].(string); ok {
			task.Command = cmd
		}
		if files, ok := m["setup_files"].(map[string]interface{}); ok {
			for path, content := range files {
				task.SetupFiles[path] = content.(string)
			}
		}
		return task
	}

	return nil
}
