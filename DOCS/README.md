# Documentazione Ufficiale: Motore "oa"

## 1. Introduzione e Filosofia
**oa** (termine dialettale per indicare *eggs*) è un motore avanzato per il remastering e la creazione di sistemi live GNU/Linux. 
L'intero progetto abbraccia la filosofia di sviluppo descritta in [Eggs & Bananas](https://penguins-eggs.net/blog/eggs-bananas): trovare soluzioni dirette, semplici e funzionali, evitando inutili complicazioni architetturali.

Maggiori infomazioni: [OA](./OA.md).

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

## Strategie
## 4. mksquashfs
* **Turbo SquashFS**: Compressione multi-core ad alte prestazioni (es. `zstd` livello 3) per generare il filesystem live in tempi minimi. 

## 5. ISO bootloader universale
* I bootloader hanno la funzione di avviare linux, ho pensato gia in penguins-eggs che per ottenere una migliore manutenzione fosse necessario prvedere una sola sorgente per i bootloader sulla ISO: grub ed isolinux. L'azione `coa-bootloaders` astrae la complessità del boot estraendo i binari necessari da quello di Debian ed installamdo solo il grub monolitico. Si perde la possibilità del secure boot, ma basta disabilitarlo per risolvere e, sul sistema installato non cambia niente.

## 6. mirror dei gruppi e gestione utenti yacto_style
Ogni distribuzione ha i propri gruppi base e, non solo, ci possono essere gruppi creati ad hoc dai customizzatori, starci appresso è impossibile, così ho ideato una nuova strategia di mirroring dei gruppi e scritto una funzione yocto_style che va a modificare direttamente i file responsabili (passwd, groups e shadow).

COA/OA aggiunge all'utente live i gruppi dell'utente SUDO_USER con una strategia mirror, prescingendo - quindi - dalle diversità presenti in ogni distribuzione.

Questo, insieme alla funzione yocto_style, permette un approccio universale agli utenti ed ai gruppi di appartenenza.

# Approfondimenti

## [ARCHITECTURE](COA_ARCHITECTURE.md)
## PACKAGE [pilot](./COA_PKG_PILOT.md)
## PACKAGE [engine](./COA_PKG_ENGINE.md)


