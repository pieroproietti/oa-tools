# Indice della Documentazione Ufficiale

Benvenuti nella documentazione di **oa-tools**. 
Questa directory contiene i dettagli architetturali, le filosofie di sviluppo e le guide di riferimento per il motore (oa) e l'orchestratore (coa).

---

## 🦾 Il Motore: "oa" (The Workhorse)
Tutto ciò che riguarda l'engine in C, la manipolazione a basso livello e l'architettura JSON-Driven.

* **[Introduzione e Filosofia (Eggs & Bananas)](./OA.md#introduzione)**: L'approccio minimalista e diretto per il remastering di sistemi live.
* **[Il "Cervello" JSON](./OA.md#architettura-json-driven)**: Come configurare l'ereditarietà dei parametri, gruppi e Initrd senza ricompilare il codice.
* **[Architettura e Sicurezza (C Engine)](./OA.md#architettura-e-sicurezza)**: Dettagli sull'isolamento (MS_PRIVATE), Zero-Copy, OverlayFS e scudo Anti-Inception.

### Strategie del Motore
* **[Compressione mksquashfs](./OA.md#mksquashfs)**: Utilizzo del Turbo SquashFS (zstd) per massimizzare le performance.
* **[ISO Bootloader Universale](./OA.md#bootloader)**: L'astrazione della complessità di avvio tramite l'azione `coa-bootloaders` (GRUB monolitico e isolinux).
* **[Gestione Utenti (yocto_style) e Gruppi](./OA.md#utenti-e-gruppi)**: La strategia di mirroring per adattarsi a qualsiasi distribuzione modificando direttamente `passwd`, `groups` e `shadow`.

---

## 🧠 L'Orchestratore: "coa" (The Mind)
Documentazione relativa al ciclo di vita, la creazione delle ISO e i package Go.

* **[Architettura Generale di COA](./COA_ARCHITECTURE.md)**: Panoramica sul funzionamento dell'orchestratore.
* **[Package: Pilot](./COA_PKG_PILOT.md)**: Dettagli e documentazione del pacchetto *pilot*.
* **[Package: Engine](./COA_PKG_ENGINE.md)**: Dettagli e documentazione del pacchetto *engine*.