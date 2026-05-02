package calamares

import (
	"coa/pkg/pilot"
	"fmt"
	"os"
	"strings"
)

// PrepareOABootloader genera i file necessari alla finalizzazione del sistema in staging (/tmp/coa)
func PrepareOABootloader(profile *pilot.Profile) error {
	// Creiamo la staging area
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return err
	}

	// 1. GENERAZIONE DI oa-bootloader.sh (Il Lavoratore nel Target)
	var worker strings.Builder
	worker.WriteString("#!/bin/bash\nset -e\n\n")
	worker.WriteString("# Rilevamento dinamico disco\n")
	worker.WriteString("TARGET_DISK=$(grub-probe -t disk / 2>/dev/null || echo \"/dev/sda\")\n\n")

	for _, step := range profile.Install {
		if step.Action == "calamares" {
			continue
		}
		cmd := strings.ReplaceAll(step.RunCommand, "/dev/sda", "$TARGET_DISK")
		worker.WriteString(fmt.Sprintf("# Step: %s\n%s\n\n", step.Name, cmd))
	}

	err := os.WriteFile(stagingDir+"/oa-bootloader.sh", []byte(worker.String()), 0755)
	if err != nil {
		return err
	}

	// 2. GENERAZIONE DI oa-prepare-target.sh (Il Ponte sulla Live)
	bridgeScript := `#!/bin/bash
TARGET_ROOT=$(mount | grep proc | grep calamares | awk '{print $3}' | sed -e "s#/proc##g")
if [ -z "$TARGET_ROOT" ]; then TARGET_ROOT=$(ls -d /tmp/calamares-root-* | head -n 1); fi

echo "OA-Bootloader: Operando su $TARGET_ROOT"
cp /tmp/coa/oa-bootloader.sh "$TARGET_ROOT/tmp/oa-bootloader.sh"
chmod +x "$TARGET_ROOT/tmp/oa-bootloader.sh"
chroot "$TARGET_ROOT" /bin/bash /tmp/oa-bootloader.sh
rm "$TARGET_ROOT/tmp/oa-bootloader.sh"
`
	err = os.WriteFile(stagingDir+"/oa-prepare-target.sh", []byte(bridgeScript), 0755)
	if err != nil {
		return err
	}

	// 3. GENERAZIONE DEL MODULO (Nella staging area)
	moduleContent := `---
dontChroot: true
timeout: 600
script:
    - "/bin/bash /tmp/coa/oa-prepare-target.sh"
`
	return os.WriteFile(stagingDir+"/shellprocess_oa_bootloader.conf", []byte(moduleContent), 0644)
}
