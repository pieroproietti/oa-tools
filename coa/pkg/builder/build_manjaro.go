package builder

import (
	"fmt"
	"os"
	"path/filepath"
)

// buildManjaroPackage genera il file PKGBUILD specifico per Manjaro Linux.
// Utilizza il set di dipendenze testato per penguins-eggs su Manjaro.
func buildManjaroPackage(projRoot, baseVer, relNum string) {
	// Definiamo il contenuto del PKGBUILD per Manjaro
	pkgbuildContent := fmt.Sprintf(`# Maintainer: Piero Proietti <piero.proietti@gmail.com>
# coa is the mind and oa the arm
pkgname=oa-tools-manjaro
pkgver=%s
pkgrel=%s
pkgdesc="oa-tools universal Linux remastering (Manjaro edition)"
arch=('x86_64')
license=('GPL3')
# Optimized Manjaro dependencies for oa-tools
depends=(
    'manjaro-tools-iso'      # Hook miso per initramfs (fondamentale su Manjaro)
    'libisoburn'             # xorriso
    'squashfs-tools'         # mksquashfs
    'mtools'                 # manipolazione EFI img
    'dosfstools'             # mkfs.vfat
    'arch-install-scripts'   # arch-chroot
    'grub'                   # bootloader
    'rsync'                  # copia file
    'sudo'                   # privilegi
    'pv'                     # progress meter
    'git'                    # gestione wardrobe
)

conflicts=('penguins-eggs')
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
    # 1. Installazione binari e creazione alias eggs
    install -Dm755 "${startdir}/oa/oa" "${pkgdir}/usr/bin/oa"
    install -Dm755 "${startdir}/coa/coa" "${pkgdir}/usr/bin/coa"
    ln -s coa "${pkgdir}/usr/bin/eggs"

    # 2. Configurazione di sistema (/etc/oa-tools.d)
    install -d "${pkgdir}/etc/oa-tools.d/brain.d"
    cp -r "${startdir}/coa/brain.d/"* "${pkgdir}/etc/oa-tools.d/brain.d/"

    # Generazione del file principale oa-tools.yaml (Dialetto oa)
    cat <<EOF > "${pkgdir}/etc/oa-tools.d/oa-tools.yaml"
---
# oa-tools configuration (Manjaro)
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

    if [ -d "${startdir}/conf" ]; then
        cp -r "${startdir}/conf/"* "${pkgdir}/etc/oa-tools.d/"
    fi

    # 3. Documentazione e Completamenti
    install -Dm644 "${startdir}/coa/docs/man/"*.1 -t "${pkgdir}/usr/share/man/man1/"
    install -Dm644 "${startdir}/coa/docs/completion/coa.bash" "${pkgdir}/usr/share/bash-completion/completions/coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.zsh" "${pkgdir}/usr/share/zsh/vendor-completions/_coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/coa.fish"

    ln -s coa "${pkgdir}/usr/share/bash-completion/completions/eggs"
    ln -s _coa "${pkgdir}/usr/share/zsh/vendor-completions/_eggs"
    ln -s coa.fish "${pkgdir}/usr/share/fish/vendor_completions.d/eggs.fish"

    echo "complete -o default -F __start_coa eggs" >> "${pkgdir}/usr/share/bash-completion/completions/coa"
}
`, baseVer, relNum)

	// Scrittura del file PKGBUILD nella root del progetto
	err := os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to write PKGBUILD: %v\n", err)
		return
	}
	fmt.Printf("[SUCCESS] PKGBUILD (Manjaro) generato correttamente per la versione %s-%s\n", baseVer, relNum)
}