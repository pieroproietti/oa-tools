package engine

import (
	"coa/pkg/pilot"
)

// OATask rappresenta un singolo comando atomico per il binario oa
type OATask struct {
	Command    string       `json:"command"`
	Info       string       `json:"info,omitempty"`
	Path       string       `json:"path,omitempty"`
	Src        string       `json:"src,omitempty"`
	Dst        string       `json:"dst,omitempty"`
	Type       string       `json:"type,omitempty"`
	Opts       string       `json:"opts,omitempty"`
	ReadOnly   bool         `json:"readonly,omitempty"`
	RunCommand string       `json:"run_command,omitempty"`
	Chroot     bool         `json:"chroot,omitempty"`
	Users      []pilot.User `json:"users,omitempty"`
	PathLiveFs string       `json:"pathLiveFs,omitempty"`
}

// OAPlan è l'array di task che oa itererà
type OAPlan struct {
	Plan []OATask `json:"plan"`
}
