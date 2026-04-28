package calamares

import (
    "coa/pkg/pilot"
    "strings"
    "os"
    "fmt"
)

func PrepareFinalizeScript(profile *pilot.Profile) error {
    os.MkdirAll("/tmp/coa", 0755)
    
    var sb strings.Builder
    sb.WriteString("#!/bin/bash\nset -e\n\n")
    sb.WriteString("# Autodetect del disco\n")
    sb.WriteString("TARGET_DISK=$(grub-probe -t disk / 2>/dev/null || echo \"/dev/sda\")\n")

    for _, step := range profile.Install {
        if step.Command == "calamares" { continue }
        
        // Sostituiamo /dev/sda con la variabile dinamica
        cmd := strings.ReplaceAll(step.Command, "/dev/sda", "$TARGET_DISK")
        sb.WriteString(fmt.Sprintf("\necho \"Step: %s\"\n%s\n", step.Name, cmd))
    }

    return os.WriteFile("/tmp/coa/finalize.sh", []byte(sb.String()), 0755)
}
