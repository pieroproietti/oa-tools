// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package engine

import (
	"os"
	"strings"
)

// generateExcludeList crea il file .list dinamico per mksquashfs
func generateExcludeList(mode string) string {
	outPath := "/tmp/coa/excludes.list"
	var excludes []string

	excludes = append(excludes,
		"boot/efi/EFI",
		"boot/loader/entries/",
		"etc/fstab",
		"var/lib/docker/",
	)

	if mode != "clone" && mode != "crypted" {
		excludes = append(excludes, "root/*")
	}

	userList := "/etc/coa/exclusion.list"
	if _, err := os.Stat(userList); os.IsNotExist(err) {
		userList = "conf/exclusion.list"
	}

	if data, err := os.ReadFile(userList); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				excludes = append(excludes, line)
			}
		}
	}

	os.MkdirAll("/tmp/coa", 0755)
	os.WriteFile(outPath, []byte(strings.Join(excludes, "\n")+"\n"), 0644)

	return outPath
}
