#!/bin/bash
# 95luks-loop: Apre luks.img e lo rende disponibile per dmsquash-live

type info >/dev/null 2>&1 || info() { echo "[luks-loop] $*" >&2; }
type warn >/dev/null 2>&1 || warn() { echo "[luks-loop][WARN] $*" >&2; }

# Leggi parametri kernel
luks_img=$(getarg rd.luks.loop=)
luks_uuid=$(getarg eggs.luks.uuid=)

info "luks-loop starting..."
info "luks_img=$luks_img"
info "luks_uuid=$luks_uuid"

# Se non specificato, prova il path scritto da iso-scan
if [ -z "$luks_img" ] && [ -e /run/initramfs/luks.loop.path ]; then
    luks_img=$(cat /run/initramfs/luks.loop.path)
    info "luks_img from iso-scan: $luks_img"
fi

# Se ancora non c'è, esci
if [ -z "$luks_img" ]; then
    info "no rd.luks.loop parameter, skipping"
    exit 0
fi

# Attendi che il file sia disponibile
info "waiting for $luks_img..."
i=0
while [ ! -f "$luks_img" ]; do
    sleep 1
    i=$((i+1))
    if [ $i -gt 30 ]; then
        warn "timeout waiting for $luks_img"
        exit 1
    fi
done

info "found $luks_img"

# Setup loop device
loop_dev=$(losetup -f --show "$luks_img")
if [ -z "$loop_dev" ]; then
    warn "losetup failed"
    exit 1
fi

info "loop device: $loop_dev"

# Attendi che udev lo riconosca
udevadm settle --timeout=10

# Apri LUKS
crypt_name="crypted"

info "opening LUKS container as $crypt_name..."

# Prova con keyfile se esiste
if [ -f /run/initramfs/live.key ]; then
    info "using keyfile"
    cryptsetup luksOpen "$loop_dev" "$crypt_name" --key-file /run/initramfs/live.key
else
    info "prompting for passphrase"
    cryptsetup luksOpen "$loop_dev" "$crypt_name"
fi

# Verifica successo
if [ ! -b "/dev/mapper/$crypt_name" ]; then
    warn "LUKS open failed!"
    losetup -d "$loop_dev"
    exit 1
fi

info "LUKS opened successfully: /dev/mapper/$crypt_name"

# Monta il filesystem decifrato in un punto temporaneo
# così dmsquash-live può trovare filesystem.squashfs dentro
mkdir -p /run/initramfs/crypted
mount -o ro /dev/mapper/$crypt_name /run/initramfs/crypted

if [ -f /run/initramfs/crypted/filesystem.squashfs ]; then
    info "found filesystem.squashfs inside LUKS"
    # Scrivi il path per dmsquash-live
    echo "/run/initramfs/crypted/filesystem.squashfs" > /run/initramfs/live.squashfs.path
else
    warn "filesystem.squashfs not found inside LUKS!"
fi

info "luks-loop completed"
exit 0