package engine

// BootloaderRoot definisce dove vengono estratti i bootloader. [cite: 88]
const BootloaderRoot = "/tmp/coa/bootloaders"

// Action rappresenta un singolo blocco "command" nell'array "plan" [cite: 89]
type Action struct {
	Command         string   `json:"command"`
	VolID           string   `json:"volid,omitempty"`
	OutputISO       string   `json:"output_iso,omitempty"`
	CryptedPassword string   `json:"crypted_password,omitempty"`
	RunCommand      string   `json:"run_command,omitempty"`
	Chroot          bool     `json:"chroot,omitempty"` // Supporto per esecuzione in liveroot
	ExcludeList     string   `json:"exclude_list,omitempty"`
	BootParams      string   `json:"boot_params,omitempty"` // Parametri dinamici per il bootloader [cite: 89]
	Args            []string `json:"args,omitempty"`
}

// UserConfig definisce la struttura per la creazione nativa dell'utente live [cite: 89]
type UserConfig struct {
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Gecos    string   `json:"gecos"`
	Home     string   `json:"home"`
	Shell    string   `json:"shell"`
	Groups   []string `json:"groups"`
}

// FlightPlan è l'oggetto JSON principale inviato al motore oa [cite: 90]
type FlightPlan struct {
	PathLiveFs      string       `json:"pathLiveFs"`
	Mode            string       `json:"mode"`
	Family          string       `json:"family"`
	InitrdCmd       string       `json:"initrd_cmd"`
	BootloadersPath string       `json:"bootloaders_path"`
	Users           []UserConfig `json:"users"`
	Plan            []Action     `json:"plan"`
}
