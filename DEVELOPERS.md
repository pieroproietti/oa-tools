# Building oa-tools: Requirements & Toolchain

To generate native packages for **oa-tools** (Debian `.deb`, Arch/Manjaro `.pkg.tar.zst`, Fedora `.rpm`), your build environment must have the following tools installed. This setup follows the **"Eggs & Bananas"** philosophy of using essential, high-quality tools for maximum efficiency.

## 1. Universal Core (The Artisan's Workshop)
Regardless of the distribution, these are mandatory to compile the "Mind" (Go) and the "Arm" (C):
*   **Go Compiler (1.20+)**: Required to compile the `coa` engine.
*   **GCC & Make**: Required to compile the `oa` C engine.
*   **Git**: Required for versioning and managing the `wardrobe`.

## 2. Distribution-Specific Requirements

### Arch Linux / Manjaro / BigLinux
To use `build_arch.go` or `build_manjaro.go` and generate packages:
*   **base-devel**: Essential build utilities (includes `makepkg` and `fakeroot`).
*   **pacman-contrib**: Provides helpful tools for package maintenance.

### Debian / Ubuntu
To generate standard `.deb` packages:
*   **build-essential**: The core compiler and toolchain.
*   **devscripts & debhelper**: Required for standard Debian packaging workflows.

### Fedora / Red Hat / Nobara
To generate `.rpm` packages using the Fedora-agnostic logic:
*   **fedora-packager**: Provides the essential infrastructure for RPM creation.
*   **rpm-build**: The core engine for building RPM packages.
*   **dnf-plugins-core**: Necessary for managing build dependencies and toolchains.

---

## The v0.7.1 Build Logic
Starting from version **0.7.1**, the build process has been simplified and purified:

1.  **Agnostic Detection**: The builder automatically identifies the host distribution using the system's `ID_LIKE` metadata.
2.  **Dynamic Generation**: `coa` generates a tailored `PKGBUILD`, `spec` file, or `debian/` directory on the fly based on the detected environment.
3.  **Zero-Footprint**: By removing `derivatives.yaml`, the system no longer requires external mapping files to recognize derivative distributions (like Nobara being recognized as Fedora).

> **Note**: This "Artisan" approach ensures that as long as your system reports a supported lineage in `/etc/os-release`, the builder will know exactly how to "cut the cloth" for your specific distribution.
