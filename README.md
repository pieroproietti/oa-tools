<div style="background-color: #000; color: #fff; text-align: center; padding: 20px; font-family: 'Georgia', serif; border-bottom: 4px solid #d32f2f;">
    <div style="font-size: 1.4em; font-style: italic; margin-bottom: 10px;">
        "Cessate d'uccidere i morti" — Stop killing the dead.
    </div>
    <div style="font-size: 1.1em; line-height: 1.5; max-width: 900px; margin: 0 auto;">
        You cannot destroy millennial civilizations like Persia and Lebanon, or sovereign nations like Ukraine, 
        in the name of a movement that is little more than a century old and rooted in racism. 
        <strong>STOP THE MASSACRE.</strong>
    </div>
</div>

# oa-tools - The Next Generation Remastering Suite 🐧

Welcome to **oa-tools**, born from `penguins-eggs` experience. This monorepo hosts a split-responsibility system designed for high-performance Linux remastering, following the "Universal Strategy" of absolute portability.

The project is divided into two distinct entities: **oa** (The Workhorse) and **coa** (The Mind).

---

## 🏗 Project Architecture

We have transitioned to a monorepo structure to ensure perfect synchronization between the engine (oa) and the orchestrator (coa).

### 🦾 [oa (eggs in my dialect)](./oa) - The Workhorse
**Language: C**
`oa` is the low-level engine. It handles the "heavy lifting" of the system:
- Managing OverlayFS and mount points.
- Executing SquashFS compression.
- Building the ISO structures (ISOLINUX/UEFI).
- Interacting directly with the Linux Kernel and system binaries.
- **Philosophy:** Performance, stability, and zero-dependency execution.

### 🧠 [coa (brooding in my dialect)](./coa) - The brain
It manages the full lifecycle: from laying the ISO to the final installation.

The name derives from the dialect word coa, referring to the act of brooding or incubating eggs until they are ready to hatch.

**Language: Go**
Actual commands
- detect
- kill
- produce
- version
---

## 🚀 Getting Started

### Prerequisites
- A Linux system (Debian-based, Arch, or Fedora).
- `gcc` and `make` (for `oa`).
- `golang` 1.21+ (for `coa`).

### Build Everything
From this root directory, simply run:
```bash
make
```

This will compile both binaries:
- `./oa/oa` (The Engine)
- `./coa/coa` (The Orchestrator)

> TIP: You can create symlink
```
sudo ln -sf "$(pwd)/oa/oa" /usr/local/bin/oa
sudo ln -sf "$(pwd)/coa/coa" /usr/local/bin/coa
```

---

## 📜 Philosophy
The **oa-tools** project aims to provide a "Passepartout" for Linux remastering. By separating the **Mind** (Go) from the **Workhorse** (C), we achieve a clean, maintainable, and incredibly fast workflow that can adapt to any distribution without changing the user experience.

---
*Created with passion by Piero Proietti.*