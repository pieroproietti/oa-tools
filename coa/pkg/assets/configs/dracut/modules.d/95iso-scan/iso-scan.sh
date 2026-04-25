#!/bin/bash
# iso-scan.sh
# Penguins-Eggs: ISO scanner for Debian Dracut

type info >/dev/null 2>&1 || info() { echo "[iso-scan] $*" >&2; }
type warn >/dev/null 2>&1 || warn() { echo "[iso-scan][WARN] $*" >&2; }
type die >/dev/null 2>&1 || die() { echo "[iso-scan][ERR] $*" >&2; exit 1; }

info "iso-scan starting..."

# Retrieve kernel args
# boot=live root=live:LABEL=colibri
# Estrai solo il LABEL, non "live:LABEL="
root_arg=$(getarg root=)
live_label=""

if [[ "$root_arg" =~ live:LABEL=(.+) ]]; then
    live_label="${BASH_REMATCH[1]}"
fi

live_media=$(getarg rd.live.media=)
live_dir=$(getarg rd.live.dir=)
[ -z "$live_dir" ] && live_dir="/live"

luks_loop=$(getarg rd.luks.loop=)
squashimg=$(getarg rd.live.squashimg=)
[ -z "$squashimg" ] && squashimg="filesystem.squashfs"

mountpoint="/run/initramfs/isoscan"
mkdir -p "$mountpoint"

info "Looking for device with LABEL=$live_label"

# Detect ISO device
device=""
if [ -n "$live_label" ]; then
    # Cerca device per LABEL
    info "searching by LABEL: $live_label"
    device=$(blkid -l -t LABEL="$live_label" -o device)
    if [ -z "$device" ]; then
        # Fallback: cerca con findfs
        device=$(findfs LABEL="$live_label" 2>/dev/null)
    fi
elif [ -n "$live_media" ]; then
    device="$live_media"
    info "using rd.live.media: $device"
else
    # Fallback: prova /dev/sr0 o loop devices
    info "no LABEL specified, trying default devices"
    for dev in /dev/sr0 /dev/sda /dev/sdb /dev/loop*; do
        if [ -b "$dev" ]; then
            device="$dev"
            info "trying device: $device"
            break
        fi
    done
fi

if [ -z "$device" ]; then
    warn "no live device found"
    exit 0
fi

info "found live device: $device"

# Monta il device
if ! mount -o ro "$device" "$mountpoint" 2>/dev/null; then
    warn "cannot mount $device on $mountpoint"
    exit 0
fi

info "mounted $device on $mountpoint"

# Verifica contenuto
ls -la "$mountpoint$live_dir/" 2>/dev/null | head -10 | while read line; do
    info "  $line"
done

# Se luks.img esiste, salva il path
if [ -e "$mountpoint$live_dir/luks.img" ]; then
    echo "$mountpoint$live_dir/luks.img" > /run/initramfs/luks.loop.path
    info "found luks.img at $mountpoint$live_dir/luks.img"
fi

# Se filesystem.squashfs esiste (senza LUKS), salva il path
if [ -e "$mountpoint$live_dir/$squashimg" ]; then
    echo "$mountpoint$live_dir/$squashimg" > /run/initramfs/live.squashfs.path
    info "found $squashimg at $mountpoint$live_dir/$squashimg"
fi

info "iso-scan completed"
exit 0
