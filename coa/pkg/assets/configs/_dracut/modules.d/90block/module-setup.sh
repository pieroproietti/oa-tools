#!/bin/bash

check() {
    return 0
}

depends() {
    # Nessuna dipendenza - questo è solo un shim
    return 0
}

install() {
    inst_multiple mount umount blkid
    inst_hook cmdline 10 "$moddir/block-cmdline.sh"
}
