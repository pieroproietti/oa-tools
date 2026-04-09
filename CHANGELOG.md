# Changelog - oa-tools

## [0.6.0] - 2026-04-09

### 🚀 Added
- **Universal Agnostic Partitioning**: The C engine now always initializes target disks with a 3-partition GPT table (`BIOSBOOT`, `EFI`, `ROOT`), ensuring universal compatibility and disk portability regardless of the host firmware.
- **Legacy BIOS Support**: Implemented the `hatch_bios.c` action to execute `grub-install --target=i386-pc` directly onto the physical disk.
- **Dynamic Routing (Krill)**: The Go orchestrator dynamically detects the presence of `/sys/firmware/efi` and injects the correct bootloader installation action (`hatch_uefi` or `hatch_bios`) into the JSON flight plan.
- **Dynamic Squashfs Detection**: The orchestrator (The Mind) automatically locates the pristine `filesystem.squashfs` image based on standard live distro mount points (e.g., `/run/live/medium/...`) and passes the exact path to the C engine.

### 🛠 Changed
- **Pristine Install Architecture (Unpack)**: Replaced `rsync` with `unsquashfs` in `hatch_unpack.c`. The installer no longer clones the actively running (and potentially polluted) Live system, but extracts the pure factory image for a clean, pristine installation.
- **Unpack Progress Bar**: Restored the standard output of `unsquashfs` during the hatch phase to display the native progress bar to the user.
- **EFI Image Size**: Increased the `efi.img` file size for hybrid ISO generation from 4MB to 10MB to safely accommodate larger modern bootloaders.
- **FAT16 Compliance**: Forced the EFI image for hybrid boot into the `-F 16` format to satisfy strict validation checks by modern UEFI firmwares (such as OVMF on Proxmox).

### 🐛 Fixed
- **Hybrid ISO Boot**: Fixed UEFI boot failures on the generated ISO. Added the `-append_partition 2 0xef` flag to `xorriso` to physically inject the EFI partition into the ISO's GPT table, making it universally bootable as both a CD-ROM and a USB drive (Raw Disk).
- **Removable Media Fallback**: Corrected the UEFI bootloader path inside the ISO to strictly comply with the removable media standard (`/EFI/BOOT/bootx64.efi`).
- **Partition Syncing**: Fixed a bug in `hatch_partition.c` where the BIOSBOOT partition (2MB) creation was skipped, which caused an index shift and subsequent formatting failures.

## [0.5.1] - 2026-04-08

### [coa] Orchestrator (The Mind)
- **New Build System**: Added `coa build` to automate C/Go compilation and native packaging (`.deb`, `PKGBUILD`).
- **Dynamic Exclusions**: Now generates runtime exclusion lists for SquashFS, moving logic away from the C engine.
- **Auto-Docs**: Integrated `coa docs` to generate man pages and shell completions (Bash, Zsh, Fish).
- **Version Controller**: Established as the Single Source of Truth for the entire project versioning.

### [oa] Engine (The Body)
- **Decoupled Logic**: Removed hardcoded exclusion arrays in `lay_squash.c` in favor of dynamic `-ef` flags.
- **Version Injection**: The engine now receives its version string during compilation via Makefile macros.
- **Streamlined Execution**: Focused the C core on pure system actions, delegating policy-making to the orchestrator.

### [General]
- **Git Hygiene**: Optimized `.gitignore` to keep the repository clean from build artifacts.
- **Single Source of Truth**: Centralized versioning in `build.go`.

## [v0.5.0] - 2026-04-02

### 🚀 Added
- **Smart Surgical Unmount**: `action_cleanup` now dynamically reads `/proc/mounts` to perform an inside-out unmount (`MNT_DETACH`). It specifically targets only `liveroot` and `.overlay` branches, eliminating "zombie mounts" safely without affecting user network shares or external drives.
- **Absolute Path Support for ISO**: `action_iso` now detects if the `output_iso` parameter is an absolute path (starts with `/`). This allows the engine to write the final ISO directly to external mounts (like a remote NAS or USB drive) bypassing local storage entirely, removing the need for `.mnt` hidden folders or symlinks.
- **Agnostic Bootloader Injection**: Introduced the `bootloaders_path` JSON parameter. `action_uefi` and `action_isolinux` can now dynamically pull bootloader binaries (like Debian's `grubx64.efi` and `isolinux.bin`) from an external path provided by the orchestrator, enabling universal booting for non-Debian host systems.
- **Hybrid UEFI/BIOS ISO Master**: `action_iso` now automatically detects the presence of UEFI payloads and dynamically generates a 4MB FAT `efi.img` (via `dd`, `mkfs.vfat`, and loop mount) to inject into `xorriso` using the `-eltorito-alt-boot` flag.
- **UEFI Trampoline Configuration**: Added a secondary `grub.cfg` inside the `EFI/BOOT/` directory to act as a trampoline. This redirects the firmware to the main `/boot/grub/grub.cfg` on the ISO9660 filesystem, completely avoiding the dreaded `grub rescue>` prompt.

### 🛠 Changed
- **Single Responsibility Principle Enforced**: Completely removed legacy, hardcoded cleanup routines from `action_prepare.c`. Now `prepare` only mounts, and `cleanup` only unmounts.
- **Plan JSONs Updated**: All JSON templates (`plan-standard.json`, `plan-clone.json`, `plan-crypted.json`) have been updated to include the `bootloaders_path` key and the correct modern modular actions.
- **Code Refactoring**: Removed the legacy monolithic `action_remaster` logic completely, distributing its responsibilities natively into `action_livestruct`, `action_isolinux`, and `action_uefi`.

### 🐛 Fixed
- **Missing GRUB Modules**: Fixed an issue in `action_uefi` where `*.mod` and `*.lst` files were not copied to the `x86_64-efi` directory, causing boot failures.

### 📚 Documented
- **Universal Strategy Manifesto**: Created `docs/UNIVERSAL_STRATEGY.md` outlining the 4 pillars of the *penguins-eggs* ecosystem (Debian Passepartout, Yocto-style identity, Initramfs abstraction, and the Orchestrator role).
- **Architecture Guide Update**: Added details about the `tmpfs` Anti-Inception shield and dynamic `/home` handling to `docs/ARCHITECTURE.md`.
- **Actions Reference**: Completely overhauled `docs/ACTIONS.md` to reflect the new C engine modules (added `cleanup`, `crypted`, `scan`, `suspend`).
- **README and Roadmap**: Cleaned up `README.md`, moved future goals to a dedicated `docs/ROADMAP.md`, and added links to the philosophical architecture.

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