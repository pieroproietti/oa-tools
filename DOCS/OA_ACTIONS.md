# 🐧 oa: Action Reference Manual

Every operation in **oa** is driven by a JSON "Plan." 
Actions are now semantically divided into three biological phases:
1. **`lay_` (Deposition/Remastering)**: Actions used to create the live ISO (the Egg).
2. **`hatch_` (Hatching/Installation)**: Actions used to install the system to physical hardware.
3. **`sys_` (System)**: Generic utility actions.

---

## 🥚 Phase 1: Lay (Remastering)

### `lay_cleanup`
**Purpose**: Safely unmounts all projections and tears down the environment.
* **Parameters**: `pathLiveFs` (String)
* **Behavior**: Detaches the `tmpfs` Anti-Recursion mask and safely unmounts all bind-mounts and OverlayFS directories using `MNT_DETACH`.

### `lay_crypted`
**Purpose**: Encapsulates the compressed filesystem into a LUKS2 encrypted container.
* **Parameters**: `pathLiveFs` (String), `crypted_password` (String)
* **Behavior**: Allocates a `root.img` file, formats it as LUKS2, creates an `ext4` filesystem inside, and moves `filesystem.squashfs` into it.

### `lay_initrd`
**Purpose**: Generates the Initial RAM Disk for the live session via template substitution.
* **Parameters**: `pathLiveFs` (String), `initrd_cmd` (String)
* **Behavior**: Replaces `{{out}}` and `{{ver}}` in the command template and triggers the ramdisk generation.

### `lay_iso`
**Purpose**: Masters the final bootable ISO image.
* **Parameters**: `pathLiveFs` (String), `volid` (String), `output_iso` (String)
* **Behavior**: Constructs the `xorriso` command, configures hybrid boot, and writes the output file.

### `lay_isolinux`
**Purpose**: Populates legacy BIOS bootloader binaries and configuration.
* **Parameters**: `pathLiveFs` (String), `bootloaders_path` (String)
* **Behavior**: Copies `isolinux.bin` and BIOS modules, generating a default boot menu.

### `lay_livestruct`
**Purpose**: Prepares the core live directory structure and extracts the host kernel.
* **Parameters**: `pathLiveFs` (String)
* **Behavior**: Detects the host's running kernel and copies the corresponding `vmlinuz`.

### `lay_prepare`
**Purpose**: Initializes the Zero-Copy environment using OverlayFS and bind mounts.
* **Parameters**: `pathLiveFs` (String), `mode` (String)
* **Behavior**: Creates the `liveroot`, bind-mounts host data read-only, sets up OverlayFS for `/usr` and `/var`, and applies the `tmpfs` Anti-Recursion mask.

### `lay_squash`
**Purpose**: Compresses the `liveroot` into a high-performance SquashFS image.
* **Parameters**: `pathLiveFs`, `compression`, `compression_level`, `exclude_list`, `mode`
* **Behavior**: Applies session exclusions and compresses using available CPU cores.

### `lay_uefi`
**Purpose**: Prepares the directory structure for UEFI booting on the ISO.
* **Parameters**: `pathLiveFs` (String), `bootloaders_path` (String)
* **Behavior**: Extracts EFI payloads and GRUB modules to `iso/EFI/BOOT`.

### `lay_users`
**Purpose**: Creates the Live user identity for the ISO.
* **Parameters**: `pathLiveFs`, `users` (Array), `mode` (String)
* **Behavior**: Purges host identities in standard mode, handcrafts the live user natively via C streams, and populates the home directory.

---

## 🐣 Phase 2: Hatch (Installation)

### `hatch_partition`
**Purpose**: Zaps the target physical disk and creates the necessary EFI and ROOT partitions.
* **Parameters**: `run_command` (String - Target disk, e.g., `/dev/sda`).

### `hatch_format`
**Purpose**: Formats the physical partitions for the new installation.
* **Parameters**: `run_command` (String - Target disk).
* **Behavior**: Formats partition 1 as `FAT32` (EFI) and partition 2 as `EXT4` (Root).

### `hatch_unpack`
**Purpose**: Physically copies the live system to the target disk.
* **Parameters**: `run_command` (String - Target disk), `pathLiveFs` (Target mount point).
* **Behavior**: Mounts the new partitions and uses `rsync` to transfer the filesystem, excluding virtual mounts.

### `hatch_fstab`
**Purpose**: Generates the `/etc/fstab` for the installed system.
* **Parameters**: `run_command` (String - Target disk), `pathLiveFs` (Target mount point).
* **Behavior**: Retrieves the partition UUIDs via `blkid` and writes the fstab natively.

### `hatch_users`
**Purpose**: Injects the new machine owner identity onto the physical disk.
* **Parameters**: `pathLiveFs` (Target mount point), `users` (Array).
* **Behavior**: Operates directly on the target's `/etc/passwd` and `/etc/shadow` without purging, injecting the user and configuring their home directory natively.

### `hatch_uefi`
**Purpose**: Installs the GRUB bootloader to the physical disk.
* **Parameters**: `run_command` (String - Target disk), `pathLiveFs` (Target mount point).
* **Behavior**: Performs API bind mounts and executes `grub-install` and `grub-mkconfig` inside the chrooted target environment.

---

## ⚙️ Phase 3: Sys (Utilities)

### `sys_run`
**Purpose**: Safely executes commands inside the chroot environment.
* **Parameters**: `pathLiveFs`, `run_command`, `args`
* **Behavior**: Forks, chroots into `pathLiveFs`, and executes via `execvp`.

### `sys_scan`
**Purpose**: Scans a specific path to calculate the total size and file count.
* **Parameters**: `path` (String)
* **Behavior**: Utilizes POSIX `nftw` for high-performance scanning.

### `sys_suspend`
**Purpose**: Pauses the engine to allow for manual inspection.
* **Parameters**: `pathLiveFs`, `message`
* **Behavior**: Halts execution until the user presses ENTER.