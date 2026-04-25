package pilot

type Profile struct {
	Remaster []YamlStep `yaml:"remaster"`
	Install  []YamlStep `yaml:"install"`
}

type YamlStep struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
	RunCommand  string `yaml:"run_command"`
	Chroot      bool   `yaml:"chroot"`
	Path        string `yaml:"path"`
	Src         string `yaml:"src"`
	Dst         string `yaml:"dst"`
}

type User struct {
	Login    string   `yaml:"login" json:"login"`
	Password string   `yaml:"password" json:"password"`
	Home     string   `yaml:"home" json:"home"`
	Shell    string   `yaml:"shell" json:"shell"`
	Groups   []string `yaml:"groups" json:"groups"`
	UID      int      `yaml:"uid" json:"uid"`
	GID      int      `yaml:"gid" json:"gid"`
}
