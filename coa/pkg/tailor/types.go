package tailor

import "coa/pkg/engine"

// WardrobeInfo per List e Show rapido
type WardrobeInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Suit rappresenta il nuovo standard index.yaml
type Suit struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Packages    []string        `yaml:"packages"`    // Pacchetti di riferimento (Debian)
	Accessories []string        `yaml:"accessories"` // Altri vestiti inclusi
	Cmds        []string        `yaml:"cmds"`        // Comandi post-install
	Dress       []engine.OATask `yaml:"dress"`       // Task complessi opzionali
}