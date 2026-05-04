# ⚙️ L'Ingegnere: `engine` (Traduzione e Compilazione del Piano)

Il package `engine` è il ponte vitale tra l'astrazione dello YAML (letta dal `pilot`) e le syscall crudo del binario C (`oa`). La sua funzione principale è prendere le intenzioni dell'utente e compilarle in un formato JSON rigoroso, agendo da "compilatore" JIT (Just-In-Time) prima dell'esecuzione.

Questo package non esegue comandi sul sistema: *prepara il terreno* affinché il motore C possa farlo senza errori.

---

## 🏗️ 1. L'Espansione Logica: `expandMountLogic()`

Quando il Pilot incontra l'azione astratta `oa_mount_logic` nello YAML, l'Engine entra in azione[cite: 27]. Questa singola direttiva viene esplosa dinamicamente in una complessa sequenza di mount point di basso livello[cite: 27]:

1.  **Setup Struttura:** Genera i task `oa_mkdir` per creare l'albero di staging (`liveroot`, `upperdir`, `lowerdir`, ecc.)[cite: 27].
2.  **Copie Fisiche:** Emette task `oa_cp` per clonare interi alberi vitali dell'host come `/etc` e `/boot`[cite: 27]. Gestisce anche la copia dinamica dei symlink del kernel (es. `vmlinuz`, `initrd.img`)[cite: 27].
3.  **Il Fix Usrmerge:** Per le directory di base (`/bin`, `/sbin`, `/lib`), analizza il sistema host con `os.Lstat`. Se rileva che la directory è in realtà un *symlink* (scenario Usrmerge, comune in Debian/Ubuntu), emette un task `oa_shell` per replicare il link. Se è una cartella vera, emette un `oa_bind`[cite: 27].
4.  **OverlayFS Dinamico:** Per `/usr` e `/var`, istruisce il motore C a costruire un filesystem Overlay complesso, montando la directory host come `lowerdir` (sola lettura) e fornendo un `upperdir` e `workdir` per la scrivibilità nell'ambiente live[cite: 27].
5.  **API FS:** Mappa i mount virtuali essenziali (`proc`, `sys`, `dev`, `run`)[cite: 27].
6.  **Chroot Fix (/tmp):** Manda un comando specifico per montare `/tmp` come `tmpfs` in RAM, forzando i permessi `1777` (Sticky Bit) per garantire la sicurezza e la compatibilità nel chroot[cite: 27].

---

## 🧹 2. La Sicurezza dei Dati: `GenerateExcludeList()`

Questa funzione è il "filtro privacy e performance"[cite: 28]. Genera il file `/tmp/coa/excludes.list` che dirà al compressore `mksquashfs` cosa *non* includere nella ISO finale[cite: 28]:

*   **Il "Doppio Colpo":** Esclude le API virtuali in due passaggi (es. `run/*` e `run/.*`) per assicurarsi che nessun file nascosto dell'host venga inavvertitamente copiato[cite: 28]. Un target primario è `var/tmp/.*` per eliminare pesanti file temporanei nascosti[cite: 28].
*   **Cache e Rete:** Rimuove le cache dei package manager (`apt/*.bin`, `pacman/pkg/*`, `dnf/*`) e informazioni di rete sensibili (`NetworkManager/system-connections/*`, chiavi SSH host)[cite: 28].
*   **Privacy Utente (Mode):** Se il parametro `--mode` è `standard`, rasa al suolo la cartella `root/*`[cite: 28]. Se è `clone`, salva i dati dell'utente ma epura comunque le cronologie (`.bash_history`, `.zsh_history`) e i cestini (`.local/share/Trash/*`)[cite: 28].
*   **Esclusioni Custom:** Se l'utente ha fornito un file `/etc/oa-tools.d/exclusion.list`, lo legge, pulisce i path e aggiunge le regole personalizzate al file di output[cite: 28].

---

## ⚙️ 3. Il Compilatore JIT: `GeneratePlan()`

È la funzione "main" dell'Engine. Accetta l'array di struct `pilot.Step` e lo converte in un JSON salvandolo in `/tmp/coa/oa-plan.json`[cite: 29].

### A. La Risoluzione delle Variabili
L'Engine agisce da template engine, sostituendo "al volo" variabili come `${ISO_OUTPUT}` con il percorso reale calcolato in Go, e `${ISO_NAME}` con il nome file del target[cite: 29].

### B. Gestione delle Azioni Complesse
*   **`oa_users`:** L'Engine prepara il task `oa_shell` per copiare lo scheletro dell'utente da `/etc/skel`[cite: 29]. Poi, legge gli utenti passati dallo YAML, recupera i gruppi dell'utente host tramite `utils.GetUserGroups()` e li inietta nel profilo live[cite: 29]. Se lo YAML non definisce utenti, usa un "salvagente" predefinito (utente `live` con password cifrata)[cite: 29].
*   **Il Breakpoint (`stopAfter`):** Se è attivo un breakpoint, l'Engine cicla fino al task richiesto, imposta `hitBreakpoint = true` e da quel momento *scarta* tutti i task successivi[cite: 29]. Tuttavia, fa un'eccezione vitale: inserisce sempre e comunque il task `coa-cleanup` alla fine, per garantire che l'host non rimanga impiccato da mountpoint orfani[cite: 29].

### C. La Magia dell'Embedding (Il package `types`)
Nel file `types.go`, l'Engine definisce la `OATask`[cite: 30]. Invece di riscrivere tutti i campi, usa l'embedding di Go (`pilot.Step json:",inline"`). Questo significa che il task "eredita" automaticamente l'azione, il comando e gli utenti dallo YAML originale[cite: 30], a cui l'Engine aggiunge solo i campi tecnici necessari al binario C (come `Type`, `Opts` e `PathLiveFs`)[cite: 30].

Questa struttura JSON, una volta salvata tramite `savePlan()`, è il pacchetto perfetto che il binario C caricherà ed eseguirà alla lettera.