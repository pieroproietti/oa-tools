# 🧠 L'Architetto Go: `coa` (Orchestratore e CLI)

Se il motore C (`oa`) è il braccio meccanico che esegue fisicamente le syscall sul sistema, il binario Go **`coa`** è l'architetto. Analizza l'ambiente, legge le regole dal Brain (i file YAML), disegna il piano di esecuzione e impartisce gli ordini.

Tutta l'interfaccia a riga di comando (CLI) è costruita sfruttando il framework **Cobra**[cite: 22]. Questo approccio garantisce modularità, autocompletamento e una gestione rigorosa dei privilegi di sistema (`CheckSudoRequirements`)[cite: 22].

---

## 🎛️ La Tabella di Comando (Il package `cmd`)

Il package `cmd` contiene l'interfaccia utente. Ogni file mappa un comando specifico che l'utente può invocare:

| Comando | Sudo Richiesto | Ruolo e Funzionamento |
| :--- | :--- | :--- |
| `remaster` | Sì | Il cuore del sistema. Avvia il "volo" per la produzione della ISO. Invoca il `pilot` per leggere le regole YAML e l'`engine` per scrivere il piano JSON, per poi lanciare il binario C[cite: 21]. Supporta parametri cruciali come `--mode` (standard, clone, crypted) e `--stop-after` per il debugging chirurgico[cite: 21]. |
| `kill` | Sì | Il distruttore sicuro. Lancia prima il motore C con `oa cleanup` per smontare i filesystem virtuali in modalità `MNT_DETACH` (per non bloccare il sistema host), poi rimuove fisicamente la workspace `/home/eggs` e i file di log[cite: 20]. |
| `detect` | No | Strumento di diagnostica in sola lettura. Utilizza il package `distro` per identificare la famiglia dell'host (es. "debian", "archlinux"), l'ID e la release, fornendo un output a colori senza alterare alcun file[cite: 15]. |
| `adapt` | No | Utility post-boot per macchine virtuali. Itera sulle uscite video virtuali (es. `Virtual-0`, `Virtual-1`) e lancia `xrandr --auto` in modo silenzioso per adattare istantaneamente la risoluzione della Live ISO alla finestra dell'hypervisor[cite: 14]. |
| `export` | No | Orchestratore di rete. Suddiviso nei sotto-comandi `iso` e `pkg`, permette di inviare (tramite `scp` e multiplexing SSH) gli artefatti compilati a uno storage Proxmox remoto[cite: 16, 17, 18]. Il flag globale `--clean` permette di cancellare le vecchie versioni sul server prima dell'upload[cite: 16, 17, 18]. |
| `sysinstall` | Misto | Comando padre che funge da router verso l'installatore finale. Permette di lanciare la GUI (`calamares`) o la TUI (`krill`) sul sistema target per finalizzare l'installazione su disco[cite: 23]. |
| `_gen_docs` | No | Comando *nascosto* (hidden). Usato in fase di build per autogenerare la documentazione Markdown, le pagine Man e gli script di autocompletamento (Bash, Zsh, Fish)[cite: 19]. |

---

## 🧬 La Dinamica di "Remaster": Pilot, Engine e C

Il capolavoro architetturale si esprime nel file `remaster.go`[cite: 21]. Quando l'utente lancia `sudo coa remaster`, il codice Go innesca una reazione a catena orchestrata magistralmente:

### 1. Riconoscimento e Setup
Inizialmente, `coa` usa `distro.NewDistro()` per capire su quale base sta atterrando (Debian, Arch, Fedora, ecc.) e calcola il percorso di output finale dell'uovo (la ISO)[cite: 21].

### 2. L'intervento del Pilot (`pilot.DetectAndLoad()`)
Il **Pilot** entra in gioco. Il suo compito è caricare lo "spartito" dal Brain[cite: 21]. Analizza il file YAML specifico per la distribuzione rilevata, verificando la correttezza della sequenza di task (come `coa-initrd`, `coa-bootloaders`, ecc.).

### 3. Il traduttore JSON: L'Engine (`engine.GeneratePlan()`)
Qui avviene il passaggio di consegne logico. L'**Engine** di Go prende l'albero YAML validato dal Pilot e lo converte in un *piano di volo JSON*[cite: 21]. 
*   **La magia del Breakpoint:** Se l'utente ha passato il flag `--stop-after` (es. `--stop-after coa-initrd`), l'Engine taglia letteralmente il piano JSON a quello step[cite: 21]. Questo permette al sistema di fermarsi a metà esecuzione, lasciando l'ambiente *chroot* montato e pronto per un'ispezione manuale e il debug[cite: 21].
*   **Preparazione Asset:** Sempre in questa fase, Go assicura la presenza dei binari dei bootloader (`utils.EnsureBootloaders`) e genera i file di esclusione dinamici (`engine.GenerateExcludeList`) in base al parametro `--mode`[cite: 21].

### 4. L'handoff al motore C (`oa`)
A questo punto, Go ha terminato il lavoro "di intelligenza" e ha prodotto un piano perfetto e sicuro (il file JSON in `/tmp/coa/`). 
Usa `exec.Command("oa", planPath)` per invocare il motore a basso livello[cite: 21]. 
Go collega lo `Stdout` e lo `Stderr` del processo C direttamente al terminale dell'utente[cite: 21], permettendo di vedere l'esecuzione in tempo reale. Se il C crasha, Go intercetta l'exit code e ferma tutto emettendo un log di errore rosso; se va a buon fine, chiude con il messaggio verde di successo[cite: 21].

---

Questo design bifasico (Logica e Astrazione in **Go** ➔ Esecuzione Syscall cruda in **C**) è ciò che rende la piattaforma **oa-tools** un ibrido perfetto: massimizza la manutenibilità e la flessibilità moderna di Go, fondendola con l'isolamento chirurgico, la velocità e la sicurezza a basso livello del linguaggio C.