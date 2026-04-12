# Distro-Specific Challenges & Edge Cases

This document tracks the technical "traps" and solutions encountered while porting **oa** across different GNU/Linux families.

## 1. Initrd Generation (The "Forge" Problem)
* **Debian/Devuan**: Uses `mkinitramfs`. Reliable, but requires `live-boot` and `live-boot-initramfs-tools` to be pre-installed on the host for a successful live boot.
* **Arch/Manjaro**: Uses `mkinitcpio`. **Problem**: It relies on specific hooks (`archiso`). **Solution**: The `skeleton` action must allow passing custom hook parameters via JSON.
* **Fedora/Suse**: Uses `dracut`. **Problem**: Highly modular and can produce very large images if not tuned.

## 2. User Groups & Permissions
* **The "Sudo" Trap**: Not all distros use the `sudo` group for administrative privileges. 
    * *Debian*: `sudo`
    * *Arch/RedHat*: `wheel`
* **Solution**: **oa** must never hardcode group names. They must be injected via the `groups[]` array in `action_users`.

## 3. Mount Propagation (The OverlayFS Ghost)
* **Problem**: When using OverlayFS, some mount events in the `liveroot` might "leak" back to the host if propagation is not handled correctly.
* **Solution**: Use `MS_PRIVATE` during the `prepare` phase to isolate the namespace surgically.

## 4. Kernel Symlinks
* **Problem**: Debian uses `/vmlinuz` as a symlink in root. Arch stores everything in `/boot/vmlinuz-linux`.
* **Solution**: **oa** should use `cp -L` (dereference) to follow symlinks regardless of their location, provided the path is passed correctly in the `plan.json`.
