# oa 🥚🥚

oa is a high-performance core engine written in C, designed for GNU/Linux system remastering. It replaces fragile and slow Bash scripting with the precision and power of native Linux kernel syscalls.

Designed as a **standalone engine** to power **penguins-eggs** and other remastering tools like **MX-Snapshot** — the real ancestor of penguins-eggs (*) — oa provides a clean, JSON-based interface to manage critical system-level operations.


## 🚀 Key Features

* **Turbo SquashFS**: High-performance compression engine with automatic multi-core detection (`-processors`) and Zstd level tuning for maximum speed.

* **Independent ISO Builder**: A standalone "Skeleton" engine inspired by the *AdrianTM* philosophy—extracting bootloaders and kernels directly from the host system without external dependencies.

* **Zero-Copy Filesystem Layering**: oa eliminates the need for physical data duplication by projecting the host hierarchy (`/bin`, `/etc`, `/usr`, `/var`...) via Read-Only mounts. It utilizes **OverlayFS** to provide a writable environment for remastering without touching the underlying host data. Critical kernel interfaces (`/dev`, `/proc`, `/sys`, `/run`) are safely bind-mounted with private propagation (`MS_PRIVATE`) for a secure, zero-footprint operation.

* **Smart Exclusions**: Supports complex exclusion lists with intelligent branch skipping (`FTW_SKIP_SUBTREE`) for maximum efficiency.

* **Zero Dependencies**: Built with a minimalist philosophy. It only requires the [cJSON](https://github.com/DaveGamble/cJSON) library (included) and standard POSIX libraries.

* **JSON-Driven**: Every action is defined by a JSON task file, making it trivial to integrate with Node.js, Python, or C++/Qt orchestrators.



## 🛠 Prerequisites & Compilation



oa expects `squashfs-tools` and `xorriso` to be available on the host system. To compile:



```bash

make

```



## 📂 Plan Execution



oa can execute complex workflows through a `plan.json`. This allows for a full remastering cycle in one shot:



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

      "volume_id": "OA_LIVE", 

      "filename": "live-system.iso" 

    },

    { "command": "action_cleanup" }

  ]

}

```



**Execution:**

```bash

sudo ./oa plan.json

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


# The name and other stuff


The name `oa` is a tribute to the linguistic roots of Zagarolo, my native town. 

Following the Sack of Rome in 1527, the local population was decimated. To repopulate the area, families were brought in from the North-East.

This created a unique linguistic pocket: while Rome's dialect evolved toward a Tuscan style, Zagarolo preserved an "Umbrian-style" dialect. Even today, the kinship with the Terni dialect is striking—sharing the same sounds, cadences, and archaic forms.

`oa` is the local word for eggs. It represents the project's evolution: moving from the broad concept of penguins-eggs to a core engine that is as essential, native, and "raw" as its name.


When I started penguins-eggs, I explored the landscape of the time, studying Systemback and Remastersys, before finally discovering refracta-snapshot, which I used as my starting point. For years, I believed mx-snapshot was a sort of "cousin" based on Refracta. However, about 10 years later, I discovered a deeper truth directly from Adrian himself: mx-snapshot was the true ancestor. I had believed it was based on refracta-snapshot, but it was actually the other way around.

Most importantly, I later re-adopted the core concept that had been set aside in Refracta: the avoidance of copying the live running filesystem. I achieved this by implementing it through a different, modern path: OverlayFS.


For those interested in the philosophy behind the project, I recommend this [article](https://penguins-eggs.net/blog/eggs-bananas) on my blog.