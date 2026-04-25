#!/bin/bash

check() {
    return 0
}

depends() {
    return 0
}

# RIMUOVI QUESTA FUNZIONE COMPLETAMENTE
# cmdline() { ... }

install() {
    inst_multiple losetup cryptsetup blkid udevadm mount
    inst_hook pre-mount 50 "$moddir/luks-loop.sh"
}