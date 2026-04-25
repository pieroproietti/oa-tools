package builder

import (
	"fmt"
	"os"
	"path/filepath"
)

// buildArchPackage genera il file PKGBUILD necessario per creare il pacchetto su Arch Linux.
func buildArchPackage(projRoot, baseVer, relNum string) {
	pkgbuildContent := fmt.Sprintf(`# Maintainer: Piero Proietti <piero.proietti@gmail.com>
pkgname=oa-tools
pkgver=%s
pkgrel=%s
pkgdesc="oa-tools universal Linux remastering"
arch=('x86_64')
license=('GPL3')
depends=('archiso' 'xorriso' 'squashfs-tools')
conflicts=('penguins-eggs')
options=(!debug)

package() {
    # 1. Installazione binari e creazione alias eggs
    install -Dm755 "${startdir}/oa/oa" "${pkgdir}/usr/local/bin/oa"
    install -Dm755 "${startdir}/coa/coa" "${pkgdir}/usr/local/bin/coa"
    ln -s coa "${pkgdir}/usr/local/bin/eggs"

    # 2. Documentazione (Man pages)
    install -Dm644 "${startdir}/coa/docs/man/"*.1 -t "${pkgdir}/usr/share/man/man1/"

    # 3. Completamenti shell e relativi alias
    install -Dm644 "${startdir}/coa/docs/completion/coa.bash" "${pkgdir}/usr/share/bash-completion/completions/coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.zsh" "${pkgdir}/usr/share/zsh/vendor-completions/_coa"
    install -Dm644 "${startdir}/coa/docs/completion/coa.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/coa.fish"

    ln -s coa "${pkgdir}/usr/share/bash-completion/completions/eggs"
    ln -s _coa "${pkgdir}/usr/share/zsh/vendor-completions/_eggs"
    ln -s coa.fish "${pkgdir}/usr/share/fish/vendor_completions.d/eggs.fish"

    # 4. FIX: Patch per l'autocompletamento Bash
    echo "complete -o default -F __start_coa eggs" >> "${pkgdir}/usr/share/bash-completion/completions/coa"
}
`, baseVer, relNum)

	err := os.WriteFile(filepath.Join(projRoot, "PKGBUILD"), []byte(pkgbuildContent), 0644)
	if err != nil {
		fmt.Printf("\033[1;31m[ERROR]\033[0m Failed to write PKGBUILD: %v\n", err)
		return
	}
	fmt.Printf("\033[1;32m[SUCCESS]\033[0m PKGBUILD generato: %s-%s\n", baseVer, relNum)
}
