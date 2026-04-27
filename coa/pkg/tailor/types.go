package tailor

import "coa/pkg/engine"

// WardrobeInfo serve per il parsing veloce di List e Show
type WardrobeInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Distro      string `yaml:"distro"`
}

// Suit corrisponde alla IMateria del tuo codice TS
type Suit struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Distro      string `yaml:"distro"`
	Reboot      bool   `yaml:"reboot"`
	Sequence    struct {
		Repositories struct {
			Update       bool     `yaml:"update"`
			Upgrade      bool     `yaml:"upgrade"`
			SourcesList  []string `yaml:"sources_list"`
			SourcesListD []string `yaml:"sources_list_d"`
		} `yaml:"repositories"`
		Packages       []string `yaml:"packages"`
		PackagesPython []string `yaml:"packages_python"`
		Cmds           []string `yaml:"cmds"`
		Accessories    []string `yaml:"accessories"`
	} `yaml:"sequence"`
	Finalize struct {
		Customize bool     `yaml:"customize"`
		Cmds      []string `yaml:"cmds"`
	} `yaml:"finalize"`
	Dress []engine.OATask `yaml:"dress"`
}
