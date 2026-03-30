# Changelog - `oa` (Output Artisan) Project

All notable changes to this project will be documented in this file.

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