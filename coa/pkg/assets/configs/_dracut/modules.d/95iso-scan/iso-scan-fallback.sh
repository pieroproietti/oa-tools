#!/bin/bash
# iso-scan-fallback.sh
# Penguins-Eggs: fallback hook to help dmsquash-live on Debian

# CREA la directory se non esiste
mkdir -p /run/initramfs/cmdline.d

if [ -e /run/initramfs/live.squashfs.path ] && [ -z "$live_squash" ]; then
    live_squash=$(cat /run/initramfs/live.squashfs.path)
    info "iso-scan-fallback: live_squash set to $live_squash"
    echo "rd.live.squashimg=$live_squash" >> /run/initramfs/cmdline.d/90-iso-scan.conf
fi
