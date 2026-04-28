# Documentazione Ufficiale: Motore "oa"

## 1. Introduzione e Filosofia
**oa** (termine dialettale per indicare *eggs*) è un motore avanzato per il remastering e la creazione di sistemi live GNU/Linux. 
L'intero progetto abbraccia la filosofia di sviluppo descritta in [Eggs & Bananas](https://penguins-eggs.net/blog/eggs-bananas): trovare soluzioni dirette, semplici e funzionali, evitando inutili complicazioni architetturali.

## 2. Il "Cervello" JSON: Configurare invece di Codificare
Il vero punto di forza di oa è la sua architettura **JSON-Driven**. 
Per portare il motore su una nuova distribuzione (che sia basata su Debian, Arch, Fedora o Suse) **non è necessario modificare o ricompilare il codice sorgente in C**. 

Il motore legge un file JSON che agisce da vero e proprio "cervello" direttivo. Attraverso un doppio puntatore (`OA_Context`), l'orchestratore gestisce l'ereditarietà dei parametri, permettendo di definire via configurazione ogni comportamento distro-specifico:
* **Gruppi e Privilegi**: L'iniezione degli utenti non ha nomi "hardcoded" nel motore. La differenza tra il gruppo di amministrazione `sudo` (usato da Debian) e `wheel` (usato da Arch/RedHat) si risolve semplicemente passando l'array di stringhe corretto nel JSON all'azione `action_users`.
* **Generazione Initrd (The "Forge" Problem)**: Che la distro usi `mkinitramfs`, `mkinitcpio` (con i suoi hook specifici) o `dracut`, i parametri custom e le logiche di esecuzione vengono passati dinamicamente allo `skeleton` direttamente dal file di configurazione.
* **Kernel Symlinks**: Gestione flessibile dei collegamenti del kernel, come il classico `/vmlinuz` alla radice utilizzato storicamente da Debian.

## 3. Architettura e Sicurezza (C Engine)
Sotto il cofano guidato dal JSON, il codice C implementa funzionalità di sistema a basso livello per garantire prestazioni e isolamento totale dal sistema ospite:
* **Zero-Copy & OverlayFS**: Proietta il filesystem host senza duplicare fisicamente i dati.
* **Isolamento Chirurgico**: Utilizza mount bind con `MS_PRIVATE` durante la preparazione per evitare che gli eventi di mount nel `liveroot` si propaghino nel sistema host.
* **Gestione Utenti Nativa**: Tramite funzioni in C dedicate (es. `yocto_add_user_to_groups`), il motore modifica direttamente i file di sistema per iniettare le identità, senza dipendere in alcun modo dai binari dell'host.
* **Anti-Inception Shield**: Un mascheramento globale tramite `tmpfs` previene dinamicamente loop di scansione ricorsivi.

## 4. Boot e Creazione ISO
* **Turbo SquashFS**: Compressione multi-core ad alte prestazioni (es. `zstd` livello 3) per generare il filesystem live in tempi minimi.
* **Gestione UEFI Universale**: L'azione `action_uefi` astrae la complessità del boot estraendo a caldo i binari necessari (`grubx64.efi`, `bootx64.efi`) e generando dinamicamente il `grub.cfg` da passare a `xorriso` per il pacchetto finale.

## 5. Sviluppi Futuri
La roadmap attuale punta a snellire ulteriormente l'integrazione:
* Implementazione del **Secure Boot** (attualmente disabilitato per garantire l'avvio base).
* Estensione del supporto nativo a `dracut` e `mkinitcpio` nell'azione `action_initrd`.
* Ottimizzazione progressiva del motore di scansione filesystem (`action_scan`).

## 6. Riferimento Azioni JSON

In questa sezione vengono documentate le singole azioni configurabili tramite il "cervello" JSON. Ogni azione definisce il comportamento specifico che il motore in C andrà ad eseguire sul sistema.

### 6.1 `action_users`
Questa azione si occupa dell'iniezione nativa degli utenti e dei loro privilegi sul sistema live, senza fare affidamento sui binari dell'host (come `useradd` o `usermod`).

**Parametri Chiave:**
* **`username`**: Il nome dell'utente live (es. `"live"` o `"oa"`).
* **`password`**: La password associata (spesso lasciata vuota o predefinita per i sistemi live).
* **`groups`**: Un array di stringhe che definisce i gruppi secondari a cui l'utente deve appartenere.

**Esempio pratico (Il problema `sudo` vs `wheel`):**
Invece di avere logiche complesse nel codice C, la differenza dei permessi di amministrazione tra distribuzioni si risolve nel JSON. 

*Configurazione per derivate Debian:*
```json
"action_users": {
  "username": "oa",
  "groups": ["sudo", "audio", "video", "plugdev"]
}
```

*Configurazione per derivate Arch/RedHat:*
```json
"action_users": {
  "username": "oa",
  "groups": ["wheel", "audio", "video", "input"]
}
```
Il motore C, tramite la funzione `yocto_add_user_to_groups`, leggerà questo array e modificherà direttamente il file `/etc/group` del filesystem isolato, garantendo una compatibilità universale.
