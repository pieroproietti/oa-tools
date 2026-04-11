#!/bin/bash

check() {
    return 0
}

depends() {
    return 0
}

install() {
    inst_multiple mount umount blkid findfs losetup
    inst_hook pre-mount 10 "$moddir/iso-scan.sh"

    # Serve a scrivere /run/initramfs/live.squashfs.path
    # prima che dmsquash-live entri in azione.
    inst_hook pre-mount 20 "$moddir/iso-scan-fallback.sh"
}
