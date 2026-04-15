package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func buildFedoraPackage(projRoot, oaDir, coaDir, baseVer, relNum string) {
	pkgName := fmt.Sprintf("oa-tools-%s-%s", baseVer, relNum)
	buildRoot := filepath.Join("/tmp", "rpmbuild_"+pkgName)

	// Pulizia e creazione struttura standard RPM
	os.RemoveAll(buildRoot)
	dirs := []string{"BUILD", "BUILDROOT", "RPMS", "SOURCES", "SPECS", "SRPMS"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(buildRoot, d), 0755)
	}

	specPath := filepath.Join(buildRoot, "SPECS", "oa-tools.spec")

	// Il blocco spec[cite: 38, 39].
	// Usa macro %{nil} per saltare l'estrazione debug sui binari Go/C.
	specContent := fmt.Sprintf(`%%define debug_package %%{nil}

Name:           oa-tools
Version:        %s
Release:        %s%%{?dist}
Summary:        coa is the mind and oa the arm
License:        GPLv3
URL:            https://penguins-eggs.net/blog/eggs-bananas

Requires:       squashfs-tools, xorriso, dosfstools, mtools, dracut-live, gdisk
Conflicts:      penguins-eggs

%%description
oa-tools universal Linux remastering. 
oa is the dialect version for eggs.

%%prep
# Nessuna preparazione, usiamo i binari compilati dal Go builder

%%build
# Nessuna build, binari gia pronti

%%install
rm -rf %%{buildroot}
mkdir -p %%{buildroot}/usr/local/bin
mkdir -p %%{buildroot}/usr/share/man/man1
mkdir -p %%{buildroot}/usr/share/bash-completion/completions
mkdir -p %%{buildroot}/usr/share/zsh/vendor-completions
mkdir -p %%{buildroot}/usr/share/fish/vendor_completions.d

# 1. Copia dei binari e creazione alias eggs
install -m 0755 %s/oa %%{buildroot}/usr/local/bin/oa
install -m 0755 %s/coa %%{buildroot}/usr/local/bin/coa
ln -s coa %%{buildroot}/usr/local/bin/eggs

# 2. Documentazione
cp %s/docs/man/*.1 %%{buildroot}/usr/share/man/man1/
gzip -9 %%{buildroot}/usr/share/man/man1/*.1

# 3. Completamenti e relativi symlink
install -m 0644 %s/docs/completion/coa.bash %%{buildroot}/usr/share/bash-completion/completions/coa
install -m 0644 %s/docs/completion/coa.zsh %%{buildroot}/usr/share/zsh/vendor-completions/_coa
install -m 0644 %s/docs/completion/coa.fish %%{buildroot}/usr/share/fish/vendor_completions.d/coa.fish

ln -s coa %%{buildroot}/usr/share/bash-completion/completions/eggs
ln -s _coa %%{buildroot}/usr/share/zsh/vendor-completions/_eggs
ln -s coa.fish %%{buildroot}/usr/share/fish/vendor_completions.d/eggs.fish

# 4. FIX: Patch per l'autocompletamento Bash
echo "complete -o default -F __start_coa eggs" >> %%{buildroot}/usr/share/bash-completion/completions/coa

%%files
/usr/local/bin/oa
/usr/local/bin/coa
/usr/local/bin/eggs
/usr/share/man/man1/*
/usr/share/bash-completion/completions/coa
/usr/share/bash-completion/completions/eggs
/usr/share/zsh/vendor-completions/_coa
/usr/share/zsh/vendor-completions/_eggs
/usr/share/fish/vendor_completions.d/coa.fish
/usr/share/fish/vendor_completions.d/eggs.fish
`, baseVer, relNum, oaDir, coaDir, coaDir, coaDir, coaDir, coaDir)

	os.WriteFile(specPath, []byte(specContent), 0644)

	fmt.Println("\033[1;34m[build]\033[0m Packing .rpm archive...")

	// Eseguiamo rpmbuild specificando la cartella temporanea come _topdir [cite: 40]
	rpmCmd := exec.Command("rpmbuild", "-bb",
		"--define", fmt.Sprintf("_topdir %s", buildRoot),
		specPath)
	rpmCmd.Stdout, rpmCmd.Stderr = os.Stdout, os.Stderr

	if err := rpmCmd.Run(); err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m RPM build failed: %v\n", err)
		return
	}

	// Recuperiamo il file generato in RPMS/x86_64 [cite: 40]
	rpmDir := filepath.Join(buildRoot, "RPMS", "x86_64")
	rpmFiles, err := filepath.Glob(filepath.Join(rpmDir, "*.rpm"))

	if err == nil && len(rpmFiles) > 0 {
		for _, file := range rpmFiles {
			finalTarget := filepath.Join(projRoot, filepath.Base(file))
			copyFile(file, finalTarget)
			fmt.Printf("\033[1;32m[SUCCESS]\033[0m Package created: \033[1m%s\033[0m\n", finalTarget)
		}
	} else {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Could not find generated RPM package in %s\n", rpmDir)
	}

	// Pulizia
	os.RemoveAll(buildRoot)
}
