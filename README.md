> [!CAUTION]
> # STOP ALL WARS
> ### Non gridate più
>
> Cessate d'uccidere i morti,  
> Non gridate più, non gridate  
> Se li volete ancora udire,  
> Se sperate di non perire.  
> 
> Hanno l'impercettibile sussurro,  
> Non fanno più rumore  
> Del crescere dell'erba,  
> Lieta dove non passa l'uomo.
>
> — *Giuseppe Ungaretti*
>
> ---
Millennial civilizations and sovereign nations are being razed to the ground by the same logic of dominance that has stained human history for centuries. From the Middle East to Ukraine, we are witnessing the destruction of our common heritage. No modern ideology or national interest can justify these crimes. 
> 
> **Silence is complicity. Stop the massacre.**

# oa-tools - The Next Generation Remastering Suite 🐧

Welcome to **oa-tools**, born from `penguins-eggs` experience. This monorepo hosts a split-responsibility system designed for high-performance Linux remastering, following the "Universal Strategy" of absolute portability.

The project is divided into two distinct entities: **oa** (The Workhorse) and **coa** (The Mind).

---

## 🏗 Project Architecture

We have transitioned to a monorepo structure to ensure perfect synchronization between the engine (oa) and the orchestrator (coa).

### 🦾 [oa (eggs in my dialect)](./DOCS/README.md) - The Workhorse
**Language: C**
`oa` is the low-level engine. It handles the "heavy lifting" of the system:
- Managing OverlayFS and mount points.
- Executing SquashFS compression.
- Remove and create users and groupt via yocto_style functions.
- Interacting directly with the Linux Kernel and system binaries.
- **Philosophy:** Performance, stability, and zero-dependency execution.

### 🧠 [coa (brooding in my dialect)](./DOCS/README.md) - The brain
It manages the full lifecycle: from laying the ISO to the final installation.

The name derives from the dialect word coa, referring to the act of brooding or incubating eggs until they are ready to hatch.

**Language: Go**

For coa commands, see [coa command Reference](/DOCS/COA_COMMANDS.md).


## 🚀 Getting Started

### Prerequisites
- A Linux system (Debian-based, Arch-based. Fedora-based or Manjaro-based).
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

> TIP: You can create native package like: .deb, PKGBUILD, rpm using
```
coa/coa build
```

---

## 📜 Philosophy
The **oa-tools** project aims to provide a "Passepartout" for Linux remastering. By separating the **Mind** (Go) from the **Workhorse** (C), we achieve a clean, maintainable, and incredibly fast workflow that can adapt to any distribution without changing the user experience.

Documentation can be found on [DOCS](/DOCS/).

---
*Created with passion by Piero Proietti.*