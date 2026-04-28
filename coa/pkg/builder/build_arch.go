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
pkgname=oa-tools
pkgver=%s
pkgrel=%s
pkgdesc="oa-tools universal Linux remastering"
arch=('x86_64')
license=('GPL3')
# Aggiunte git e rsync come dipendenze vitali per il wardrobe e la sysroot
depends=('archiso' 'xorriso' 'squashfs-tools' 'git' 'rsync' 'sudo')
conflicts=('penguins-eggs')
options=(!debug)

build() {
    # 0. Compilazione del motore C (oa)
    # Assicuriamo che il "braccio" sia compilato nativamente per l'architettura target
    msg2 "Compilazione del motore C (oa)..."
    cd "${startdir}/oa"
    make clean && make
}

package() {
    # 1. Installazione binari e creazione alias eggs
    # Arch segue rigorosamente la gerarchia /usr/bin
    install -Dm755 "${startdir}/oa/oa" "${pkgdir}/usr/bin/oa"
    install -Dm755 "${startdir}/coa/coa" "${pkgdir}/usr/bin/coa"
    ln -s coa "${pkgdir}/usr/bin/eggs"

    # 2. Configurazione di sistema (/etc/oa-tools.d)
    # Creiamo la struttura per la configurazione YAML e la brain.d per coa
    install -d "${pkgdir}/etc/oa-tools.d/brain.d"

    # Generazione del file di configurazione principale oa-tools.yaml
    # Utilizziamo il dialetto "oa" [cite: 30-03-2026] e citiamo la filosofia [cite: 29-03-2026]
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

    # Se esistono file di configurazione aggiuntivi nella cartella conf del progetto, li installiamo
    if [ -d "${startdir}/conf" ]; then
        cp -r "${startdir}/conf/"* "${pkgdir}/etc/oa-tools.d/"
    fi

    # 3. Documentazione (Man pages)
    install -Dm644 "${startdir}/coa/docs/man/"*.1 -t "${pkgdir}/usr/share/man/man1/"

    # 4. Completamenti shell e relativi alias
    install -Dm644 "${startdir}/coa/docs/completion/coa.bash" "${pkgdir}/usr/share/bash-completion/completions/coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.zsh" "${pkgdir}/usr/share/zsh/vendor-completions/_coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/coa.fish"

    ln -s coa "${pkgdir}/usr/share/bash-completion/completions/eggs"
    ln -s _coa "${pkgdir}/usr/share/zsh/vendor-completions/_eggs"
    ln -s coa.fish "${pkgdir}/usr/share/fish/vendor_completions.d/eggs.fish"

    # 5. Patch per l'autocompletamento Bash dell'alias eggs
    echo "complete -o default -F __start_coa eggs" >> "${pkgdir}/usr/share/bash-completion/completions/coa"
}
`, baseVer, relNum)

	// Scrittura del file PKGBUILD nella root del progetto
	err := os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	if err != nil {
		fmt.Printf("%s[ERROR]%s Failed to write PKGBUILD: %v\n", ColorRed, ColorReset, err)
		return
	}
	fmt.Printf("%s[SUCCESS]%s PKGBUILD generato correttamente per la versione %s-%s\n", ColorGreen, ColorReset, baseVer, relNum)
}
