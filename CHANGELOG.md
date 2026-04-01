# Changelog - `oa` (Output Artisan) Project

All notable changes to this project will be documented in this file.

## [0.2.0] - 2026-04-01

### Added
- **Anti-Recursion Shield (Inception Fix)**: Implemented a global `tmpfs` mask in `action_prepare.c` to hide the working directory (`pathLiveFs`) from `mksquashfs` and `nftw`. This definitively prevents infinite filesystem loops when the workspace is located inside a bind-mounted host directory (like `/home`).
- **Native Group Injection**: Added the `yocto_add_user_to_groups` helper in `oa-yocto.c` to natively append the live user to secondary groups (e.g., `sudo`, `cdrom`) directly into `/etc/group`, completely bypassing host binaries.
- **Skeleton Population**: `action_users` now correctly populates the live user's home directory by copying hidden configuration files from `/etc/skel` and applying recursive ownership.

### Changed
- **Smart `/home` Handling**: Refactored `action_prepare.c` to handle the `/home` directory dynamically based on the execution `mode`. It is now mounted read-only for `clone` and `crypted` modes, but created as an empty directory for `standard` mode to host the newly injected live user.
- **Cleanup Fortification**: Updated `action_cleanup` to safely unmount the new anti-recursion masks and the `/home` directory using `umount2(..., MNT_DETACH)`.

### Fixed
- **Missing Live Home in ISO**: Fixed a bug in `action_squash.c` where the `home/*` exclusion was aggressively deleting the freshly created live user's home during the `mksquashfs` compression in `standard` mode.

## [0.1.0] - 2026-03-30

### Added
- contest.sh create a context.XXX.txt to restore GEMINI contest;
- **Modular Architecture**: Decoupled core logic into `src/actions/`, making the project significantly more scalable and maintainable.
- **Execution Engine**: Implemented the `execute_verb` dispatch system in `main.c` to process the `plan.json` workflow.
- TODO: **Parameter Inheritance**: Added dual-pointer support (`cJSON *root` and `cJSON *task`) allowing actions to access both Global settings and Local task overrides.
- **Dynamic Initrd Action**: Support for command templates using `{{out}}` and `{{ver}}` placeholders for flexible initramfs generation.
- TODO: **System User Discovery**: Added `action_users` to scan for "human" users (UID >= 1000) using the POSIX `getpwent()` API.

### Changed
- **Buffer Hardening**: Introduced `PATH_SAFE` (8192) and `CMD_MAX` (32768) constants in `oa.h` to ensure safety during complex `system()` command construction.
- **Centralized Definitions**: Consolidated all system headers and function prototypes into a single master header (`include/oa.h`).
- **Mount Fortification**: Refactored `action_prepare` to use `MS_PRIVATE` bind mounts and improved OverlayFS directory structures.

### Fixed
- **Warning Purgatory**: Resolved all GCC warnings related to `-Wformat-truncation` and `-Wunused-parameter`, resulting in a "Zen" build.
- **Logic Correction**: Fixed the Megabyte calculation in `action_scan.c` by replacing the incorrect `PATH_MAX` divisor with a proper `1024*1024` calculation.
- **JSON Key Sincronization**: Fixed a bug in `action_iso.c` where the ISO filename and Volume ID were ignored due to a mismatch between JSON keys and C code.

---
*"Code is poetry, but the changelog is the history."*