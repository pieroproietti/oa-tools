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

	// Pulizia e creazione struttura standard RPM (BUILD, SPECS, RPMS, ecc.)
	os.RemoveAll(buildRoot)
	dirs := []string{"BUILD", "BUILDROOT", "RPMS", "SOURCES", "SPECS", "SRPMS"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(buildRoot, d), 0755)
	}

	specPath := filepath.Join(buildRoot, "SPECS", "oa-tools.spec")

	// Generazione dello SPEC file.
	// Nota per il porting: Fedora usa 'dracut-live' invece di 'live-boot'.
	specContent := fmt.Sprintf(`%%define debug_package %%{nil}

Name:           oa-tools
Version:        %s
Release:        %s%%{?dist}
Summary:        coa is the mind and oa the arm
License:        GPLv3
URL:            https://penguins-eggs.net/blog/eggs-bananas

# Dipendenze vitali: aggiunti git e rsync per il wardrobe e la sysroot.
Requires:       squashfs-tools, xorriso, dosfstools, mtools, dracut-live, gdisk, git, rsync, sudo
Conflicts:      penguins-eggs

%%description
oa-tools universal Linux remastering. 
coa is the mind (Go) and oa is the arm (C).
oa is the dialect version for eggs [cite: 2026-03-30].

%%prep
# I binari sono già compilati dai rispettivi builder Go e C.

%%build
# Nessuna operazione richiesta qui.

%%install
rm -rf %%{buildroot}
# Creazione gerarchia di sistema
mkdir -p %%{buildroot}/usr/bin
mkdir -p %%{buildroot}/etc/oa-tools.d/brain.d
mkdir -p %%{buildroot}/usr/share/man/man1
mkdir -p %%{buildroot}/usr/share/bash-completion/completions
mkdir -p %%{buildroot}/usr/share/zsh/vendor-completions
mkdir -p %%{buildroot}/usr/share/fish/vendor_completions.d

# 1. Installazione binari e alias 'eggs'
install -m 0755 %s/oa %%{buildroot}/usr/bin/oa
install -m 0755 %s/coa %%{buildroot}/usr/bin/coa
ln -s coa %%{buildroot}/usr/bin/eggs

# 2. Configurazione di sistema (YAML)
# Definiamo l'identità del sistema e la filosofia [cite: 2026-03-29, 2026-03-30]
cat <<EOF > %%{buildroot}/etc/oa-tools.d/oa-tools.yaml
---
# oa-tools configuration
# Philosophy: https://penguins-eggs.net/blog/eggs-bananas

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

# Copia di eventuali file extra dalla cartella conf del progetto
if [ -d "%s/conf" ]; then
    cp -r %s/conf/* %%{buildroot}/etc/oa-tools.d/
fi

# 3. Documentazione (Man pages compressate)
cp %s/docs/man/*.1 %%{buildroot}/usr/share/man/man1/
gzip -9 %%{buildroot}/usr/share/man/man1/*.1

# 4. Completamenti shell e relativi alias
install -m 0644 %s/docs/completion/coa.bash %%{buildroot}/usr/share/bash-completion/completions/coa
install -m 0644 %s/docs/completion/coa.zsh %%{buildroot}/usr/share/zsh/vendor-completions/_coa
install -m 0644 %s/docs/completion/coa.fish %%{buildroot}/usr/share/fish/vendor_completions.d/coa.fish

ln -s coa %%{buildroot}/usr/share/bash-completion/completions/eggs
ln -s _coa %%{buildroot}/usr/share/zsh/vendor-completions/_eggs
ln -s coa.fish %%{buildroot}/usr/share/fish/vendor_completions.d/eggs.fish

# 5. Patch per l'autocompletamento Bash
echo "complete -o default -F __start_coa eggs" >> %%{buildroot}/usr/share/bash-completion/completions/coa

%%files
/usr/bin/oa
/usr/bin/coa
/usr/bin/eggs
%%config(noreplace) /etc/oa-tools.d/oa-tools.yaml
/etc/oa-tools.d/brain.d/
/usr/share/man/man1/*
/usr/share/bash-completion/completions/coa
/usr/share/bash-completion/completions/eggs
/usr/share/zsh/vendor-completions/_coa
/usr/share/zsh/vendor-completions/_eggs
/usr/share/fish/vendor_completions.d/coa.fish
/usr/share/fish/vendor_completions.d/eggs.fish
`, baseVer, relNum, oaDir, coaDir, baseVer, projRoot, projRoot, coaDir, coaDir, coaDir, coaDir)

	os.WriteFile(specPath, []byte(specContent), 0644)

	fmt.Printf("%s[build]%s Packing .rpm archive...\n", ColorBlue, ColorReset)

	// Esecuzione di rpmbuild con la definizione della topdir temporanea
	rpmCmd := exec.Command("rpmbuild", "-bb",
		"--define", fmt.Sprintf("_topdir %s", buildRoot),
		specPath)
	rpmCmd.Stdout, rpmCmd.Stderr = os.Stdout, os.Stderr

	if err := rpmCmd.Run(); err != nil {
		fmt.Printf("%s[ERROR]%s RPM build failed: %v\n", ColorRed, ColorReset, err)
		return
	}

	// Recupero del pacchetto generato
	rpmDir := filepath.Join(buildRoot, "RPMS", "x86_64")
	rpmFiles, err := filepath.Glob(filepath.Join(rpmDir, "*.rpm"))

	if err == nil && len(rpmFiles) > 0 {
		for _, file := range rpmFiles {
			finalTarget := filepath.Join(projRoot, filepath.Base(file))
			copyFile(file, finalTarget)
			fmt.Printf("%s[SUCCESS]%s RPM Package created: %s\n", ColorGreen, ColorReset, finalTarget)
		}
	} else {
		fmt.Printf("%s[ERROR]%s Could not find generated RPM package in %s\n", ColorRed, ColorReset, rpmDir)
	}

	// Pulizia finale
	os.RemoveAll(buildRoot)
}
