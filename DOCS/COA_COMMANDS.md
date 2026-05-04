# 🛠️ COA Command Reference

`coa` (Calamares & OA Lightweight Architect) è l'orchestratore universale per la rimasterizzazione e l'installazione del sistema. Funge da interfaccia a riga di comando (CLI) per il progetto **oa-tools**, delegando la logica complessa al *Pilot* (Go) e l'esecuzione a basso livello all'*Engine* (C).

Di seguito la reference completa dei comandi disponibili nella **v0.7.5**.

---

## 🧭 Panoramica Rapida

| Comando | Sudo | Descrizione |
| :--- | :---: | :--- |
| **`remaster`** | 🟢 Sì | Avvia la costruzione della ISO live. |
| **`sysinstall`**| 🟡 Misto | Avvia l'installatore di sistema sul target. |
| **`kill`** | 🟢 Sì | Smonta i filesystem e pulisce la workspace in sicurezza. |
| **`detect`** | 🔴 No | Rileva e mostra il profilo del sistema host. |
| **`adapt`** | 🔴 No | Adatta dinamicamente la risoluzione video in VM. |
| **`export`** | 🔴 No | Trasferisce artefatti (ISO/pacchetti) su server remoti. |

---

## 🚀 Comandi Principali

### `coa remaster`
Il cuore del sistema. Legge il profilo YAML tramite il Pilot, genera il piano JSON ed esegue il motore C per costruire l'uovo (ISO).

*   **Uso:** `sudo coa remaster [flags]`
*   **Flags:**
    *   `--mode <string>`: Modalità di produzione (`standard`, `clone`, `crypted`). Default: `standard`.
    *   `--path <string>`: Directory di lavoro. Default: `/home/eggs`.
    *   `--stop-after <step>`: **[Debug]** Ferma l'esecuzione dopo uno step specifico (es. `coa-initrd`), lasciando il *chroot* montato per ispezioni manuali.

### `coa sysinstall`
L'orchestratore per l'installazione del sistema operativo su disco. Agisce da router verso i motori di installazione finali.

*   **Uso:** `sudo coa sysinstall <engine>`
*   **Motori supportati:**
    *   `calamares`: Avvia l'interfaccia grafica (GUI).
    *   `krill`: Avvia l'interfaccia testuale (TUI).

### `coa kill`
Il "distruttore sicuro". Esegue il teardown dell'ambiente di rimasterizzazione. Utilizza un `MNT_DETACH` (lazy unmount) per liberare i mountpoint virtuali (`/proc`, `/sys`, `/dev`) senza causare kernel panic o blocchi sul sistema host, per poi eliminare la directory di lavoro.

*   **Uso:** `sudo coa kill`

---

## 🧰 Utility e Diagnostica

### `coa detect`
Strumento diagnostico in sola lettura. Analizza `/etc/os-release` per identificare l'ambiente ospite in modo agnostico e mappa la distribuzione alla famiglia madre corretta (es. riconosce *Linux Mint* come famiglia *Debian*).

*   **Uso:** `coa detect`

### `coa adapt`
Utility post-boot pensata specificamente per gli ambienti Live avviati in Macchine Virtuali (VirtualBox, QEMU/KVM, VMware). Mappa le uscite video virtuali e forza un ridimensionamento dinamico (`xrandr --auto`) per adattare la risoluzione alla finestra dell'hypervisor.

*   **Uso:** `coa adapt`

---

## 📦 Gestione Artefatti (Network)

### `coa export`
Orchestratore di rete basato su SSH Multiplexing per il trasferimento rapido e sicuro degli artefatti verso uno storage Proxmox remoto.

*   **Sotto-comandi:**
    *   `coa export iso`: Trova l'ultima ISO generata nel nido e la trasferisce.
    *   `coa export pkg`: Cerca i pacchetti nativi compilati (`.deb`, `.rpm`, `.pkg.tar.zst`) in base alla famiglia della distro e li invia.
*   **Flag globale:**
    *   `--clean`: Prima di effettuare l'upload, si collega al server e cancella le vecchie versioni dell'artefatto per risparmiare spazio sul nodo di destinazione.

---

## ⚙️ Comandi Sviluppatore (Hidden)

### `coa _gen_docs`
Comando nascosto utilizzato dalla toolchain (Makefile) durante la compilazione. Autogenera:
1.  Documentazione Markdown.
2.  Pagine Man (`man 1 coa`).
3.  Script di autocompletamento nativi per Bash, Zsh e Fish.

*   **Uso:** `coa _gen_docs --target <dir>`
