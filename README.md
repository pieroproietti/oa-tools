# Vitellus 🐧🥚

**Vitellus** (Latin for *yolk*) is a high-performance core engine written in C, designed for GNU/Linux system remastering. It was born to replace fragile and slow Bash scripting with the precision and power of native Linux kernel syscalls.

Designed as an agnostic bridge between **penguins-eggs** and othere remastering tools like **MX-Snapshot**, Vitellus provides a clean, JSON-based interface to manage critical system-level operations.

## 🚀 Key Features

* **Safe Bind Mounts**: Manages system mounts (`/dev`, `/proc`, `/sys`) using private propagation (`MS_PRIVATE`) to ensure the host system remains untouched.
* **Turbo Scan**: High-speed recursive filesystem scanning using the native `nftw` (New File Tree Walk) function.
* **Smart Exclusions**: Supports complex exclusion lists (compatible with Refracta/Eggs formats) with intelligent branch skipping (`FTW_SKIP_SUBTREE`) for maximum efficiency.
* **Zero Dependencies**: Built with a minimalist philosophy. It only requires the [cJSON](https://github.com/DaveGamble/cJSON) library (included) and standard POSIX libraries.
* **JSON-Driven**: Every action is defined by a JSON task file, making it trivial to integrate with Node.js, Python, or C++/Qt orchestrators.

## 🛠 Compilation

Vitellus is built to be lightweight. To compile it within your development VM:

```bash
gcc src/*.c -Isrc -o vitellus -lm
```

## 📂 Task Structure (Example)

Vitellus executes atomic operations based on JSON input or a plan.json file. 

For example, to prepare a work environment:

```json
{
    "command": "action_prepare",
    "pathLiveFs": "/home/eggs"
}
```

**Execution:**
```bash
sudo ./vitellus prepare.json
```
or
```bash
sudo ./vitellus plan.json
```

## 🗺 Roadmap

- [ ] Filesystem Scanning with external exclusion file support.
- [x] Secure Mount/Umount Engine.
- [ ] Creating filesystem.squashfs and iso image
- [ ] Implementation of Hooks for chroot customization.
- [ ] Direct integration into `penguins-eggs` as the primary analysis engine.

---
*Developed with the efficiency of C and the reliability of a Clipper '87 veteran.*
