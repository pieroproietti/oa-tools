package pilot

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3" // Assicurati di fare go get gopkg.in/yaml.v3
)

// DetectAndLoad trova lo spartito YAML giusto in base alla distro host
func DetectAndLoad() (*Profile, error) {
	// 1. Qui potresti usare pkg/distro per capire se caricare debian.yaml o arch.yaml
	// Per ora forziamo debian.yaml come nel tuo esempio
	brainPath := "/etc/coa/brain.d/debian.yaml"

	// Se siamo in dev, cerchiamo in locale
	if _, err := os.Stat("brain.d/debian.yaml"); err == nil {
		brainPath = "coa/brain.d/debian.yaml"
	}

	data, err := os.ReadFile(brainPath)
	if err != nil {
		return nil, fmt.Errorf("errore lettura file brain: %v", err)
	}

	var profile Profile
	err = yaml.Unmarshal(data, &profile)
	if err != nil {
		return nil, fmt.Errorf("errore parsing YAML: %v", err)
	}

	return &profile, nil
}
