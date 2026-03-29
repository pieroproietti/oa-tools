# oa Architecture: Zero-Copy Filesystem

The core philosophy of **oa** is surgical efficiency. Traditional remastering tools spend hours performing a physical copy (`cp -a`) of the host filesystem. **oa** eliminates this bottleneck by implementing a **Zero-Copy** strategy powered by the Linux Kernel's **OverlayFS**.

---

## 🏗️ The Three-Layer Logic

Instead of duplicating data, **oa** creates a "projected" filesystem by stacking distinct layers. This approach, perfected over years in the *penguins-eggs* project, is now implemented in native C for maximum performance.

[Image of OverlayFS architecture showing LowerDir, UpperDir, and MergedDir layers]

### 1. LowerDir (The Host - Read-Only)
* **Source**: The root filesystem (`/`) of the running host.
* **State**: Mounted as **Strictly Read-Only**.
* **Impact**: Zero disk I/O, zero time consumed. It serves as the immutable "DNA" of the new ISO.

### 2. UpperDir (The Workspace - Read/Write)
* **Source**: A dedicated directory in the work path (e.g., `ovfs/upper`).
* **State**: Writable.
* **Role**: This layer captures the "delta." Any modification (adding users, deleting logs, changing configs) is stored here without ever touching the host.

### 3. MergedDir (The Liveroot)
* **Role**: The unified mount point.
* **Appearance**: To the engine and tools (`chroot`, `mksquashfs`), it looks like a standard, writable root filesystem.
* **Mechanism**: The Kernel merges the Host and the Workspace in real-time.

---

## 🛠️ Implementation & Isolation

To ensure a "Zero-Footprint" operation, **oa** follows a specific execution flow during the `prepare` phase:

1.  **Mount Namespace Isolation**: Uses `MS_PRIVATE` propagation to ensure that mounts inside the `liveroot` do not "leak" back to the host system.
2.  **Kernel API Projection**: Critical interfaces are bind-mounted from the host into the `merged` directory to allow system tools to function:
    * `/dev`, `/proc`, `/sys`, `/run`.
3.  **The Overlay Syscall**: The final projection is achieved via the `mount()` syscall:
    ```c
    mount("overlay", merged_path, "overlay", 0, "lowerdir=/,upperdir=path/upper,workdir=path/work");
    ```

---

## 🚀 Key Advantages

* **Instant Setup**: The `liveroot` is ready for customization in milliseconds, regardless of the host size (10GB or 1TB).
* **SSD Longevity**: Eliminates Gigabytes of unnecessary write cycles.
* **Physical Safety**: Since the Host (LowerDir) is Read-Only, the running system is physically shielded from accidental damage during the remastering process.
* **Atomic Cleanup**: Removing the `liveroot` simply requires unmounting the overlay and deleting the small "upper" directory.

---
*Documented by Piero Proietti - 2026*