package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Helper per spostare file tra partizioni diverse
func moveFile(src, dst string) error {
	// Tenta prima il rename (veloce se sulla stessa partizione)
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Se fallisce per cross-device link, copia e rimuovi
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	sourceFile.Close()
	return os.Remove(src)
}

func buildFedoraPackage(projRoot, oaDir, coaDir, baseVer, relNum string) {
	cleanVer := strings.TrimPrefix(baseVer, "v")
	pkgName := fmt.Sprintf("oa-tools-%s-%s", cleanVer, relNum)
	buildRoot := filepath.Join("/tmp", "rpmbuild_"+pkgName)

	LogBuild("Packing .rpm archive for %s...", pkgName)

	os.RemoveAll(buildRoot)
	dirs := []string{"BUILD", "BUILDROOT", "RPMS", "SOURCES", "SPECS", "SRPMS"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(buildRoot, d), 0755)
	}

	specPath := filepath.Join(buildRoot, "SPECS", "oa-tools.spec")

	specContent := fmt.Sprintf(`%%define debug_package %%{nil}

Name:           oa-tools
Version:        %s
Release:        %s%%{?dist}
Summary:        coa is the mind and oa the arm
License:        GPLv3
URL:            https://penguins-eggs.net/blog/eggs-bananas

Requires:       squashfs-tools, xorriso, dosfstools, mtools, dracut-live, gdisk, git, rsync, sudo
Conflicts:      penguins-eggs

%%description
oa-tools universal Linux remastering. 
Evolution Edition - Dialect: oa.

%%install
rm -rf %%{buildroot}
mkdir -p %%{buildroot}/usr/bin
mkdir -p %%{buildroot}/etc/oa-tools.d/brain.d
mkdir -p %%{buildroot}/usr/share/man/man1
mkdir -p %%{buildroot}/usr/share/bash-completion/completions
mkdir -p %%{buildroot}/usr/share/zsh/vendor-completions
mkdir -p %%{buildroot}/usr/share/fish/vendor_completions.d

# Binari
install -m 0755 %s/oa/oa %%{buildroot}/usr/bin/oa
install -m 0755 %s/coa/coa %%{buildroot}/usr/bin/coa
ln -s coa %%{buildroot}/usr/bin/eggs

# Brain e Config
cp %s/coa/brain.d/*.yaml %%{buildroot}/etc/oa-tools.d/brain.d/
cat <<EOF > %%{buildroot}/etc/oa-tools.d/oa-tools.yaml
---
system:
  dialect: "oa"
  version: "%s"
wardrobe:
  root: "~/.oa-wardrobe"
  repo: "https://github.com/pieroproietti/oa-wardrobe.git"
remaster:
  default_user: "artisan"
  work_dir: "/home/eggs"
EOF

# Docs e Completions
cp %s/coa/docs/man/*.1 %%{buildroot}/usr/share/man/man1/
gzip -9 %%{buildroot}/usr/share/man/man1/*.1
install -m 0644 %s/coa/docs/completion/coa.bash %%{buildroot}/usr/share/bash-completion/completions/coa
install -m 0644 %s/coa/docs/completion/coa.zsh %%{buildroot}/usr/share/zsh/vendor-completions/_coa
install -m 0644 %s/coa/docs/completion/coa.fish %%{buildroot}/usr/share/fish/vendor_completions.d/coa.fish

ln -s coa %%{buildroot}/usr/share/bash-completion/completions/eggs
ln -s _coa %%{buildroot}/usr/share/zsh/vendor-completions/_eggs
ln -s coa.fish %%{buildroot}/usr/share/fish/vendor_completions.d/eggs.fish

%%files
/usr/bin/oa
/usr/bin/coa
/usr/bin/eggs
%%dir /etc/oa-tools.d
%%dir /etc/oa-tools.d/brain.d
%%config(noreplace) /etc/oa-tools.d/oa-tools.yaml
/etc/oa-tools.d/brain.d/*.yaml
/usr/share/man/man1/*.1.gz
/usr/share/bash-completion/completions/*
/usr/share/zsh/vendor-completions/*
/usr/share/fish/vendor_completions.d/*

%%changelog
* Fri May 01 2026 Piero Proietti <piero.proietti@gmail.com> - %s-%s
- Automatic build for Fedora Evolution Edition
`, cleanVer, relNum, projRoot, projRoot, projRoot, cleanVer, projRoot, projRoot, projRoot, projRoot, cleanVer, relNum)

	os.WriteFile(specPath, []byte(specContent), 0644)

	LogBuild("Running rpmbuild...")
	rpmCmd := exec.Command("rpmbuild", "-bb", "--define", fmt.Sprintf("_topdir %s", buildRoot), specPath)
	rpmCmd.Stdout = os.Stdout
	rpmCmd.Stderr = os.Stderr

	if err := rpmCmd.Run(); err != nil {
		LogError("RPM build failed: %v", err)
		return
	}

	rpmPattern := filepath.Join(buildRoot, "RPMS", "x86_64", "*.rpm")
	matches, _ := filepath.Glob(rpmPattern)
	if len(matches) > 0 {
		finalPkg := filepath.Join(projRoot, filepath.Base(matches[0]))
		// Usiamo moveFile invece di os.Rename per gestire partizioni diverse
		if err := moveFile(matches[0], finalPkg); err == nil {
			LogBuild("✅ RPM Package created: %s", finalPkg)
		} else {
			LogError("Failed to move RPM to root: %v", err)
		}
	} else {
		LogError("Could not find generated RPM in %s", buildRoot)
	}

	os.RemoveAll(buildRoot)
}
