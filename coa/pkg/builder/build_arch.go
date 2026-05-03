package builder

import (
	"fmt"
	"os"
	"path/filepath"
)

// buildArchPackage genera il file PKGBUILD necessario per creare il pacchetto su Arch Linux.
// Integra la configurazione YAML in /etc/oa-tools.d e la gestione del wardrobe.
func buildArchPackage(projRoot, baseVer, relNum string) {
	// Definiamo il contenuto del PKGBUILD come un template dinamico
	pkgbuildContent := fmt.Sprintf(`# Maintainer: Piero Proietti <piero.proietti@gmail.com>
# coa is the mind and oa the arm
pkgname=oa-tools-arch
pkgver=%s
pkgrel=%s
pkgdesc="oa-tools universal Linux remastering"
arch=('x86_64')
license=('GPL3')
depends=(
    'mkinitcpio-archiso'     # FONDAMENTALE per il supporto live su base Arch
    'efibootmgr'             # EFI
    'libisoburn'             # xorriso
    'squashfs-tools'         # mksquashfs
    'mtools'                 # EFI
    'dosfstools'             # vfat per EFI
    'arch-install-scripts'   # arch-chroot
    'grub'                   # bootloader
    'rsync'                  # copia file
    'sudo'                   # privilegi
    'pv'                     # progress meter
    'git'                    # gestione wardrobe
)
conflicts=('penguins-eggs' 'oa-tools')
backup=('etc/oa-tools.d/oa-tools.yaml')
options=(!debug)

build() {
    # Compilazione del "braccio" (C)
    msg2 "Compilazione del motore C (oa)..."
    cd "${startdir}/oa"
    make clean && make

    # Compilazione della "mente" (Go)
    msg2 "Compilazione del motore Go (coa)..."
    cd "${startdir}/coa"
    go build -ldflags "-X 'coa/pkg/cmd.AppVersion=${pkgver}'" -o coa main.go
}

package() {
    # 1. Installazione binari e alias
    install -Dm755 "${startdir}/oa/oa" "${pkgdir}/usr/bin/oa"
    install -Dm755 "${startdir}/coa/coa" "${pkgdir}/usr/bin/coa"
    ln -s coa "${pkgdir}/usr/bin/eggs"

    # 2. Configurazione di sistema (/etc/oa-tools.d)
    # Creiamo le directory necessarie
    install -d "${pkgdir}/etc/oa-tools.d/brain.d"

    # Copia della "mente" (index.yaml e altri file di logica)
    if [ -d "${startdir}/coa/brain.d" ]; then
        msg2 "Installazione logica agnostica (brain.d)..."
        cp -r "${startdir}/coa/brain.d/"* "${pkgdir}/etc/oa-tools.d/brain.d/"
    fi

    # Generazione del file di configurazione principale oa-tools.yaml
    cat <<EOF > "${pkgdir}/etc/oa-tools.d/oa-tools.yaml"
---
# oa-tools configuration
# Philosophy: https://penguins-eggs.net/blog/eggs-bananas

system:
  dialect: "oa"
  version: "${pkgver}"

wardrobe:
  root: "~/.oa-wardrobe"
  repo: "https://github.com/pieroproietti/oa-wardrobe.git"

remaster:
  default_user: "artisan"
  work_dir: "/home/eggs"
EOF

    # Installazione di eventuali file extra dalla cartella conf del progetto
    if [ -d "${startdir}/conf" ]; then
        msg2 "Installazione file di configurazione addizionali..."
        cp -r "${startdir}/conf/"* "${pkgdir}/etc/oa-tools.d/"
    fi

    # 3. Documentazione (Man pages)
    install -Dm644 "${startdir}/coa/docs/man/"*.1 -t "${pkgdir}/usr/share/man/man1/"

    # 4. Completamenti shell e alias
    install -Dm644 "${startdir}/coa/docs/completion/coa.bash" "${pkgdir}/usr/share/bash-completion/completions/coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.zsh" "${pkgdir}/usr/share/zsh/vendor-completions/_coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/coa.fish"

    ln -s coa "${pkgdir}/usr/share/bash-completion/completions/eggs"
    ln -s _coa "${pkgdir}/usr/share/zsh/vendor-completions/_eggs"
    ln -s coa.fish "${pkgdir}/usr/share/fish/vendor_completions.d/eggs.fish"

    # Patch autocompletamento Bash per alias eggs
    echo "complete -o default -F __start_coa eggs" >> "${pkgdir}/usr/share/bash-completion/completions/coa"
}
`, baseVer, relNum)

	// Scrittura del file PKGBUILD
	err := os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to write PKGBUILD: %v\n", err)
		return
	}
	fmt.Printf("[SUCCESS] PKGBUILD (Arch) generato correttamente per la versione %s-%s\n", baseVer, relNum)
}