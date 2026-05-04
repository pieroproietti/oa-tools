# ⚙️ Il Motore C: `oa` (Execution Engine)

Se `coa` (scritto in Go) è la mente che progetta il "piano di volo" analizzando i file YAML, il binario **`oa`** (scritto in C) è il braccio meccanico che esegue fisicamente il lavoro a basso livello. 

Il motore C riceve il piano sotto forma di un file JSON parsato tramite `cJSON`[cite: 6]. La funzione cuore del sistema è `execute_verb`, che estrae la chiave `action` da ogni task e la instrada verso il modulo C nativo competente[cite: 6]. Se incontra un'azione denominata semplicemente `shell`, la ignora demandandone la competenza al livello Go[cite: 6].

---

## 🎛️ Tabella delle Azioni Native (C-Core)

Di seguito le azioni operative mappate dal router interno `execute_verb`[cite: 6]:

| Azione JSON | Funzione C | Ruolo e Funzionamento |
| :--- | :--- | :--- |
| `oa_mkdir` | `oa_mkdir()` | Crea la directory specificata usando `mkdir -p` per garantire la catena di percorsi[cite: 7]. |
| `oa_bind` | `oa_bind()` | Esegue un bind mount (`MS_BIND | MS_REC`). Assicura l'isolamento fortificando il mount con `MS_PRIVATE` e supporta la modalità read-only (`MS_RDONLY`)[cite: 7]. |
| `oa_cp` | `oa_cp()` | Effettua copie fisiche usando `cp -a` per preservare rigorosamente permessi, symlink e timestamp originali[cite: 7]. |
| `oa_mount_generic`| `oa_mount_generic()`| Crea al volo la directory di destinazione e invoca la syscall `mount()` per filesystem virtuali (proc, sysfs, overlay)[cite: 7]. |
| `oa_shell` | `oa_shell()` | Esegue comandi shell. Può girare sull'host tramite `system()` o entrare in modalità `chroot` usando `fork()`, `chroot()` e `execl()`[cite: 9]. |
| `oa_users` | `oa_users()` | Gestisce le identità: rimuove (sanitize) gli utenti host e inietta l'utente live (Purge & Inject)[cite: 11, 12]. |
| `oa_umount` | `oa_umount()` | Legge `/proc/mounts`, individua tutto ciò che appartiene al progetto e lo smonta partendo dal percorso più profondo[cite: 10]. |

---

## 🔬 Deep Dive: I Moduli Operativi

### 1. Il Passpartout: `oa_shell`
Questo modulo è il ponte perfetto tra l'orchestratore e il sistema. Legge i parametri `run_command` e `chroot`[cite: 9].
*   **Motore Bimodale**: Se il target è il sistema da installare (`mode="install"`), il target root è `pathLiveFs`, altrimenti punta a `pathLiveFs/liveroot`[cite: 9].
*   **Chroot Nativo**: Per eseguire comandi isolati, genera un processo figlio con `fork()`, usa la syscall `chroot()` e posiziona l'ambiente su `/` con `chdir("/")` prima di lanciare `/bin/sh -c`[cite: 9]. Il processo padre attende la fine dell'esecuzione catturando l'exit code[cite: 9].

### 2. Gestione Identità: `oa_users`
Un modulo in stile Yocto Project per la sicurezza e la privacy. Opera in due fasi:
*   **Purge (Sanitize)**: Se la modalità non è "clone" o "crypted", ripulisce i file `/etc/passwd`, `/etc/shadow` e `/etc/group` dell'ambiente live rimuovendo gli ID degli utenti umani (host)[cite: 12].
*   **Inject**: Legge l'array JSON `users` e inietta le nuove identità nativamente. Utilizza `crypt()` con il salt `$6$oa$` per generare password in SHA-512 se non sono già hash[cite: 12]. Inoltre, implementa un fix per i sistemi Debian creando esplicitamente il gruppo primario (GID) per l'utente[cite: 12]. Infine, popola la home directory copiando i file da `/etc/skel`[cite: 12].

### 3. L'Infrastruttura di Mount
La magia dietro la velocità di `oa` sta nell'usare syscall C dirette al posto di script bash.
*   **Isolamento**: Il comando `oa_bind` prima esegue il mount ricorsivo, poi se richiesto ri-monta in `MS_RDONLY`, e infine lo blinda con `MS_PRIVATE`[cite: 7]. 
*   **OverlayFS**: I mount virtuali più complessi, come quelli per unire `lowerdir` e `upperdir` sulle directory `usr` e `var`, sono gestiti passingando opzioni strutturate direttamente alla syscall `mount()`[cite: 8]. La directory `/tmp` viene gestita montando un `tmpfs` con permessi rigidi `mode=1777`[cite: 8].

### 4. Smart Umount e Cleanup di Emergenza
La stabilità dell'host dipende dalla corretta pulizia.
*   **La via pulita (`oa_umount`)**: Apre `/proc/mounts`, filtra i mount point che iniziano con la root del progetto, li inserisce in un array e li ordina per lunghezza decrescente[cite: 10]. Questo garantisce che i path più annidati (es. `.../liveroot/proc`) vengano smontati prima delle directory genitore[cite: 10]. Utilizza il flag `MNT_DETACH` (lazy unmount) per forzare la chiusura senza bloccare il sistema[cite: 10].
*   **La via dura (Emergency Cleanup)**: Se l'eseguibile C viene invocato passando direttamente l'argomento `cleanup` (es. `oa cleanup`), il `main` ignora il parser JSON ed esegue una raffica rapida di `umount2(..., MNT_DETACH)` su path hardcoded fondamentali per liberare l'host istantaneamente in caso di crash fatali[cite: 6].