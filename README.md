# Vitellus 🐧🥚



**Vitellus** (Latin for *yolk*) is a high-performance core engine written in C, designed for GNU/Linux system remastering. It was born to replace fragile and slow Bash scripting with the precision and power of native Linux kernel syscalls.



Designed as a **standalone engine** to power **penguins-eggs** and other remastering tools like **MX-Snapshot** — the real father of penguins-eggs (*) — Vitellus provides a clean, JSON-based interface to manage critical system-level operations.



## 🚀 Key Features



* **Turbo SquashFS**: High-performance compression engine with automatic multi-core detection (`-processors`) and Zstd level tuning for maximum speed.

* **Independent ISO Builder**: A standalone "Skeleton" engine inspired by the *AdrianTM* philosophy—extracting bootloaders and kernels directly from the host system without external dependencies.


* **Safe Filesystem Orchestration**: Vitellus manages the entire root structure (`bin`, `etc`, `usr`, `var`...) while safely bind-mounting critical kernel interfaces (`/dev`, `/proc`, `/sys`, `/run`) using private propagation (`MS_PRIVATE`) to ensure the host system remains untouched.

* **Smart Exclusions**: Supports complex exclusion lists with intelligent branch skipping (`FTW_SKIP_SUBTREE`) for maximum efficiency.

* **Zero Dependencies**: Built with a minimalist philosophy. It only requires the [cJSON](https://github.com/DaveGamble/cJSON) library (included) and standard POSIX libraries.

* **JSON-Driven**: Every action is defined by a JSON task file, making it trivial to integrate with Node.js, Python, or C++/Qt orchestrators.



## 🛠 Prerequisites & Compilation



Vitellus expects `squashfs-tools` and `xorriso` to be available on the host system. To compile:



```bash

make

```



## 📂 Plan Execution (The "Matrimonio")



Vitellus can execute complex workflows through a `plan.json`. This allows for a full remastering cycle in one shot:



```json

{

  "pathLiveFs": "/home/eggs",

  "plan": [

    { "command": "action_prepare" },

    { "command": "action_skeleton" },

    { 

      "command": "action_squash", 

      "compression": "zstd", 

      "compression_level": 3 

    },

    { 

      "command": "action_iso", 

      "volume_id": "VITELIUS_LIVE", 

      "filename": "live-system.iso" 

    },

    { "command": "action_cleanup" }

  ]

}

```



**Execution:**

```bash

sudo ./vitellus plan.json

```



## 🗺 Roadmap



- [x] Secure Mount/Umount Engine.

- [x] Multi-core SquashFS creation.

- [x] Standalone ISO Bootloader "Skeleton" (BIOS/UEFI).

- [ ] Filesystem Scanning with external exclusion file support.

- [ ] Implementation of Hooks for chroot customization.

- [ ] Direct integration into `penguins-eggs` as the primary analysis engine.



---

*Developed with the efficiency of C and the reliability of a Clipper '87 veteran.*

> **Historical Note:** From the start of penguins-eggs, I was convinced that my work was a direct descendant of *refracta-snapshot* and really it is. However, I recently discovered—directly from Adrian himself—that *refracta-snapshot* was another derivative of his original project. l