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

### 5. Plan execution
Questo è un esempio di ciò che oa riceve da coa: un piano completo di rimasterizzazione in JSON,

Perchè JSON e non il più leggibile YAML: perchè è un piano che deve essere eseguito in c ed è molto più leggero scriverlo ed eseguirlo in JSON.

```
{
  "plan": [
    {
      "description": "Setup base path",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/liveroot"
    },
    {
      "description": "Setup base path",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay"
    },
    {
      "description": "Setup base path",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/upperdir"
    },
    {
      "description": "Setup base path",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/workdir"
    },
    {
      "description": "Setup base path",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/lowerdir"
    },
    {
      "description": "Copia fisica /etc",
      "action": "oa_cp",
      "chroot": false,
      "src": "/etc",
      "dst": "/home/eggs/liveroot"
    },
    {
      "description": "Copia fisica /boot",
      "action": "oa_cp",
      "chroot": false,
      "src": "/boot",
      "dst": "/home/eggs/liveroot"
    },
    {
      "description": "Copia symlink: vmlinuz",
      "action": "oa_cp",
      "chroot": false,
      "src": "/vmlinuz",
      "dst": "/home/eggs/liveroot/vmlinuz"
    },
    {
      "description": "Copia symlink: initrd.img",
      "action": "oa_cp",
      "chroot": false,
      "src": "/initrd.img",
      "dst": "/home/eggs/liveroot/initrd.img"
    },
    {
      "description": "Copia symlink: vmlinuz.old",
      "action": "oa_cp",
      "chroot": false,
      "src": "/vmlinuz.old",
      "dst": "/home/eggs/liveroot/vmlinuz.old"
    },
    {
      "description": "Copia symlink: initrd.img.old",
      "action": "oa_cp",
      "chroot": false,
      "src": "/initrd.img.old",
      "dst": "/home/eggs/liveroot/initrd.img.old"
    },
    {
      "description": "Replica Usrmerge symlink: bin",
      "action": "oa_shell",
      "run_command": "ln -sf usr/bin /home/eggs/liveroot/bin",
      "chroot": false
    },
    {
      "description": "Replica Usrmerge symlink: sbin",
      "action": "oa_shell",
      "run_command": "ln -sf usr/sbin /home/eggs/liveroot/sbin",
      "chroot": false
    },
    {
      "description": "Replica Usrmerge symlink: lib",
      "action": "oa_shell",
      "run_command": "ln -sf usr/lib /home/eggs/liveroot/lib",
      "chroot": false
    },
    {
      "description": "Replica Usrmerge symlink: lib64",
      "action": "oa_shell",
      "run_command": "ln -sf usr/lib64 /home/eggs/liveroot/lib64",
      "chroot": false
    },
    {
      "description": "Bind mount proiettivo: opt",
      "action": "oa_bind",
      "chroot": false,
      "src": "/opt",
      "dst": "/home/eggs/liveroot/opt",
      "readonly": true
    },
    {
      "description": "Bind mount proiettivo: root",
      "action": "oa_bind",
      "chroot": false,
      "src": "/root",
      "dst": "/home/eggs/liveroot/root",
      "readonly": true
    },
    {
      "description": "Bind mount proiettivo: srv",
      "action": "oa_bind",
      "chroot": false,
      "src": "/srv",
      "dst": "/home/eggs/liveroot/srv",
      "readonly": true
    },
    {
      "description": "",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/lowerdir/usr"
    },
    {
      "description": "",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/upperdir/usr"
    },
    {
      "description": "",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/workdir/usr"
    },
    {
      "description": "",
      "action": "oa_bind",
      "chroot": false,
      "src": "/usr",
      "dst": "/home/eggs/.overlay/lowerdir/usr",
      "readonly": true
    },
    {
      "description": "Overlay mount per scrivibilità: usr",
      "action": "oa_mount_generic",
      "chroot": false,
      "src": "overlay",
      "dst": "/home/eggs/liveroot/usr",
      "type": "overlay",
      "opts": "lowerdir=/home/eggs/.overlay/lowerdir/usr,upperdir=/home/eggs/.overlay/upperdir/usr,workdir=/home/eggs/.overlay/workdir/usr"
    },
    {
      "description": "",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/lowerdir/var"
    },
    {
      "description": "",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/upperdir/var"
    },
    {
      "description": "",
      "action": "oa_mkdir",
      "chroot": false,
      "path": "/home/eggs/.overlay/workdir/var"
    },
    {
      "description": "",
      "action": "oa_bind",
      "chroot": false,
      "src": "/var",
      "dst": "/home/eggs/.overlay/lowerdir/var",
      "readonly": true
    },
    {
      "description": "Overlay mount per scrivibilità: var",
      "action": "oa_mount_generic",
      "chroot": false,
      "src": "overlay",
      "dst": "/home/eggs/liveroot/var",
      "type": "overlay",
      "opts": "lowerdir=/home/eggs/.overlay/lowerdir/var,upperdir=/home/eggs/.overlay/upperdir/var,workdir=/home/eggs/.overlay/workdir/var"
    },
    {
      "description": "",
      "action": "oa_mount_generic",
      "chroot": false,
      "src": "proc",
      "dst": "/home/eggs/liveroot/proc",
      "type": "proc"
    },
    {
      "description": "",
      "action": "oa_mount_generic",
      "chroot": false,
      "src": "sys",
      "dst": "/home/eggs/liveroot/sys",
      "type": "sysfs"
    },
    {
      "description": "API FS: dev",
      "action": "oa_bind",
      "chroot": false,
      "src": "/dev",
      "dst": "/home/eggs/liveroot/dev"
    },
    {
      "description": "API FS: run",
      "action": "oa_bind",
      "chroot": false,
      "src": "/run",
      "dst": "/home/eggs/liveroot/run"
    },
    {
      "description": "API FS: tmp (Sticky Bit + Tmpfs)",
      "action": "oa_shell",
      "run_command": "mkdir -p /home/eggs/liveroot/tmp \u0026\u0026 chmod 1777 /home/eggs/liveroot/tmp \u0026\u0026 mount -t tmpfs -o mode=1777 tmpfs /home/eggs/liveroot/tmp",
      "chroot": false
    },
    {
      "name": "coa-bootloaders",
      "description": "Copia dei bootloader e creazione immagine FAT per UEFI",
      "action": "oa_shell",
      "run_command": "set -e\nmkdir -p /home/eggs/isodir/live\nmkdir -p /home/eggs/isodir/isolinux\nmkdir -p /home/eggs/isodir/boot/grub\nmkdir -p /home/eggs/isodir/EFI/BOOT\n\n# ISOLINUX\ncp /tmp/coa/bootloaders/ISOLINUX/isolinux.bin /home/eggs/isodir/isolinux/\ncp /tmp/coa/bootloaders/syslinux/modules/bios/*.c32 /home/eggs/isodir/isolinux/\ncp /tmp/coa/bootloaders/ISOLINUX/isolinux.cfg /home/eggs/isodir/isolinux/ 2\u003e/dev/null || touch /home/eggs/isodir/isolinux/isolinux.cfg\n\n# GRUB \u0026 UEFI\ncp /tmp/coa/bootloaders/grub/grub.cfg /home/eggs/isodir/boot/grub/ 2\u003e/dev/null || touch /home/eggs/isodir/boot/grub/grub.cfg\ncp /tmp/coa/bootloaders/grub/x86_64-efi/monolithic/grubx64.efi /home/eggs/isodir/EFI/BOOT/BOOTX64.EFI\n\n# Creazione FAT EFI IMG\ndd if=/dev/zero of=/home/eggs/isodir/EFI/BOOT/efi.img bs=1k count=8192\nmkfs.vfat /home/eggs/isodir/EFI/BOOT/efi.img\nmmd -i /home/eggs/isodir/EFI/BOOT/efi.img ::/EFI\nmmd -i /home/eggs/isodir/EFI/BOOT/efi.img ::/EFI/BOOT\nmcopy -i /home/eggs/isodir/EFI/BOOT/efi.img /home/eggs/isodir/EFI/BOOT/BOOTX64.EFI ::/EFI/BOOT/BOOTX64.EFI\nmcopy -i /home/eggs/isodir/EFI/BOOT/efi.img /home/eggs/isodir/boot/grub/grub.cfg ::/EFI/BOOT/grub.cfg\n",
      "chroot": false,
      "pathLiveFs": "/home/eggs"
    },
    {
      "description": "Creazione home directory da /etc/skel",
      "action": "oa_shell",
      "run_command": "mkdir -p /home/eggs/liveroot/home/live \u0026\u0026 cp -a /home/eggs/liveroot/etc/skel/. /home/eggs/liveroot/home/live/",
      "chroot": false
    },
    {
      "description": "Iniezione identità utenti live",
      "action": "oa_users",
      "chroot": false,
      "users": [
        {
          "login": "live",
          "password": "$6$oa-tools$uTKAYeAVn.Y.Dy2To6HXsHt1Gt4HpMghmOV93a46jFY7hkAQ3tk7eRTKjcvSYDf5sOf3qnKzyyPYXurKp9ST3.",
          "home": "/home/live",
          "shell": "/bin/bash",
          "groups": [
            "cdrom",
            "floppy",
            "sudo",
            "audio",
            "dip",
            "video",
            "plugdev",
            "netdev",
            "wheel"
          ],
          "uid": 1000,
          "gid": 1000
        }
      ],
      "pathLiveFs": "/home/eggs"
    },
    {
      "name": "coa-enable-live",
      "description": "Abilitazione poteri sudo per user live",
      "action": "oa_shell",
      "run_command": "mkdir -p /etc/sudoers.d\nrm -f /etc/sudoers.d/00*\necho \"live ALL=(ALL) NOPASSWD: ALL\" \u003e /etc/sudoers.d/00-live\nchmod 0440 /etc/sudoers.d/00-live\n",
      "chroot": true,
      "pathLiveFs": "/home/eggs"
    },
    {
      "name": "coa-initrd",
      "description": "Generazione Initramfs custom nel chroot",
      "action": "oa_shell",
      "run_command": "update-initramfs -u -k all || update-initramfs -c -k all",
      "chroot": true,
      "pathLiveFs": "/home/eggs"
    },
    {
      "name": "coa-copy-kernel",
      "description": "Copia del Kernel e Initrd direttamente da /boot",
      "action": "oa_shell",
      "run_command": "# Trova l'ultima versione del kernel e initrd in /boot\nVMLINUZ=$(ls -v /home/eggs/liveroot/boot/vmlinuz-* | tail -n 1)\nINITRD=$(ls -v /home/eggs/liveroot/boot/initrd.img-* | tail -n 1)\n\n# Esegui la copia rinominandoli per la ISO\ncp \"$VMLINUZ\" /home/eggs/isodir/live/vmlinuz\ncp \"$INITRD\" /home/eggs/isodir/live/initrd.img\n",
      "chroot": false,
      "pathLiveFs": "/home/eggs"
    },
    {
      "name": "coa-live-menus",
      "description": "Generazione configurazioni di boot con branding e splash",
      "action": "oa_shell",
      "run_command": "set -e\n# 1. Variabili di ambiente e branding\nVOL_ID=\"OA_LIVE\"\nBOOT_PARAMS=\"boot=live components quiet splash\"\nWORK_DIR=\"/home/eggs/isodir\"\n# Recuperiamo il nome reale della distro\nPRETTY=$(grep \"PRETTY_NAME\" /etc/os-release | cut -d'\"' -f2 || echo \"OA Live\")\n\n# 2. Creazione Directory\nmkdir -p ${WORK_DIR}/boot/grub\nmkdir -p ${WORK_DIR}/isolinux\nmkdir -p ${WORK_DIR}/EFI/BOOT\n\n# 3. Copia Assets (Font e Splash)\n# Copia del font Unicode per evitare i quadratini\ncp /usr/share/grub/unicode.pf2 ${WORK_DIR}/boot/grub/font.pf2 || echo \"Font non trovato, ignoro.\"\n\n# Copia dello splash screen dagli asset locali\nif [ -f \"/etc/oa-tools.d/brain.d/assets/splash.png\" ]; then\n    cp /etc/oa-tools.d/brain.d/assets/splash.png ${WORK_DIR}/boot/grub/splash.png\n    cp /etc/oa-tools.d/brain.d/assets/splash.png ${WORK_DIR}/isolinux/splash.png\nfi\n\n# 4. Generazione GRUB.cfg (Main)\ncat \u003c\u003cEOF \u003e ${WORK_DIR}/boot/grub/grub.cfg\nset timeout=5\nset default=2\n\ninsmod all_video\ninsmod gfxterm\ninsmod png\ninsmod part_gpt\ninsmod part_msdos\ninsmod fat\ninsmod iso9660\n\n# Caricamento Font e Terminale Grafico\nif loadfont /boot/grub/font.pf2; then\n    set gfxmode=auto\n    terminal_output gfxterm\nfi\n\n# Immagine di sfondo e colori (nero su trasparente per contrasto)\nbackground_image /boot/grub/splash.png\nset menu_color_normal=black/black\nset menu_color_highlight=white/black\n\nsearch --no-floppy --set=root --label ${VOL_ID}\n\n# --- ELEMENTI ESTETICI ---\n# 0. Voce Titolo (non selezionabile o comunque inerte)\nmenuentry \"--- Penguins eggs (COA/OA Edition) ---\" {\n    true\n}\n\n# 1. Voce Vuota (Separatore)\nmenuentry \"\" {\n    true\n}\n\nmenuentry \"Start $PRETTY\" {\n    linux /live/vmlinuz ${BOOT_PARAMS}\n    initrd /live/initrd.img\n}\n\nmenuentry \"Start $PRETTY - RAM mode\" {\n    linux /live/vmlinuz ${BOOT_PARAMS} toram\n    initrd /live/initrd.img\n}\nEOF\n\n# 5. Generazione ISOLINUX.cfg (BIOS)\ncat \u003c\u003cEOF \u003e ${WORK_DIR}/isolinux/isolinux.cfg\nUI vesamenu.c32\nTIMEOUT 50\nDEFAULT live\n\nMENU BACKGROUND splash.png\nMENU TITLE Penguins eggs (COA/OA edition)\n\nLABEL live\n    MENU LABEL Start $PRETTY\n    LINUX /live/vmlinuz\n    APPEND ${BOOT_PARAMS}\n    INITRD /live/initrd.img\n\nLABEL ram\n    MENU LABEL Start $PRETTY - RAM mode\n    LINUX /live/vmlinuz\n    APPEND ${BOOT_PARAMS} toram\n    INITRD /live/initrd.img\nEOF\n\n# 6. Generazione Trampolino EFI\ncat \u003c\u003cEOF \u003e ${WORK_DIR}/EFI/BOOT/grub.cfg\nsearch --set=root --file /live/filesystem.squashfs\nset prefix=(\\$root)/boot/grub\nconfigfile \\$prefix/grub.cfg\nEOF\n",
      "chroot": false,
      "pathLiveFs": "/home/eggs"
    },
    {
      "name": "coa-squashfs",
      "description": "Compressione del sistema (SquashFS zstd)",
      "action": "oa_shell",
      "run_command": "mksquashfs /home/eggs/liveroot /home/eggs/isodir/live/filesystem.squashfs -comp zstd -Xcompression-level 3 -b 1M -processors 4 -noappend -wildcards -ef /tmp/coa/excludes.list",
      "chroot": false,
      "pathLiveFs": "/home/eggs"
    },
    {
      "name": "coa-xorriso",
      "description": "Creazione della ISO finale con xorriso",
      "action": "oa_shell",
      "run_command": "xorriso -as mkisofs -iso-level 3 -R -J -joliet-long -l -cache-inodes -V 'OA_LIVE' -isohybrid-mbr /tmp/coa/bootloaders/ISOLINUX/isohdpfx.bin -partition_offset 16 -A 'OA_LIVE' -b isolinux/isolinux.bin -c isolinux/boot.cat -no-emul-boot -boot-load-size 4 -boot-info-table -eltorito-alt-boot -e EFI/BOOT/efi.img -no-emul-boot -isohybrid-gpt-basdat -o /home/eggs/egg-of-debian-trixie-colibri-amd64-2026-05-04_1152.iso /home/eggs/isodir",
      "chroot": false,
      "pathLiveFs": "/home/eggs"
    },
    {
      "description": "Pulizia finale dei mount",
      "action": "oa_umount",
      "chroot": false,
      "pathLiveFs": "/home/eggs"
    }
  ]
}
```