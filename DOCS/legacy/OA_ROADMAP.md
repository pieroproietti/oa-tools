# 🗺 oa: Project Roadmap

This document tracks the completed milestones and the upcoming goals for the **oa** engine.

## ✅ Completed Milestones
- [x] **Secure Engine Foundation**: Implementation of isolated mount/umount engine using `MS_PRIVATE` bind mounts.
- [x] **Zero-Copy Architecture**: Full integration of OverlayFS to project the host filesystem without physical duplication.
- [x] **Native Identity Injection**: Standalone C implementation to craft users and inject groups natively without host binaries.
- [x] **Turbo SquashFS**: High-performance multi-core compression integration.
- [x] **Anti-Inception Shield**: Global `tmpfs` masking to prevent recursive scanning loops dynamically.
- [x] **JSON-Driven Orchestrator**: Dual-pointer (`OA_Context`) parameter inheritance allowing global and local task overrides.
- [x] **UEFI Bootloader Completion**: Finalize `action_uefi` to hot-extract `grubx64.efi`, `bootx64.efi`, and generate `grub.cfg` for the ISO.

## 🚧 Work in Progress / Next Steps
- [x] **SECURE BOOT**: For now we must disable SECURE_BOOT.
- [ ] **Advanced Initramfs Handling**: Extend `action_initrd` to seamlessly tame `dracut` (Fedora/SUSE) and `mkinitcpio` (Arch Linux).
- [ ] **Filesystem Scanning Engine**: Refine `action_scan` and `action_squash` to support dynamic, external exclusion lists effectively.
