package engine

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// --- HELPER DI SISTEMA ---
func isUEFI() bool {
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		return true
	}
	return false
}

func getSquashfsPath() string {
	paths := []string{
		"/run/live/medium/live/filesystem.squashfs",       // Debian Live
		"/lib/live/mount/medium/live/filesystem.squashfs", // Debian Alt
		"/run/archiso/bootmnt/arch/x86_64/airootfs.sfs",   // Arch
		"/run/initramfs/live/live/filesystem.squashfs",    // Fedora
		"/home/eggs/iso/live/filesystem.squashfs",         // Local coa/oa
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func getRootDisk() string {
	cmd := "lsblk -n -o PKNAME $(findmnt -n -v -o SOURCE /) | head -n 1"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getAvailableDisks() []string {
	rootDisk := getRootDisk()
	out, _ := exec.Command("lsblk", "-d", "-n", "-o", "NAME,SIZE,MODEL").Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var disks []string

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			diskName := parts[0]
			if diskName == rootDisk ||
				strings.HasPrefix(diskName, "loop") ||
				strings.HasPrefix(diskName, "zram") ||
				strings.HasPrefix(diskName, "sr") {
				continue
			}
			diskStr := fmt.Sprintf("/dev/%s - %s %s", diskName, parts[1], strings.Join(parts[2:], " "))
			disks = append(disks, diskStr)
		}
	}
	if len(disks) == 0 {
		disks = append(disks, "NO_SAFE_DISKS_FOUND")
	}
	return disks
}

func generateHashedPassword(plain string) string {
	out, err := exec.Command("openssl", "passwd", "-6", plain).Output()
	if err != nil {
		return plain
	}
	return strings.TrimSpace(string(out))
}
