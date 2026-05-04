# 🧠 brain.d/YAML_LOGIC.md (Evolution Edition - Pro Level)

Questo documento rappresenta la documentazione tecnica avanzata degli script YAML utilizzati dal **Pilot** per l'automazione della build in **oa-tools**. Ogni script è disegnato per tradurre il sistema operativo host in una distribuzione live installabile (l'uovo), mantenendo una logica universale (il dialetto **oa**) e adattandosi dinamicamente alle specificità di Debian, Arch, Manjaro e Fedora.

---

## 🛠️ PARTE 1: La Fase di Remaster (Costruzione dell'Uovo)

La sezione `remaster:` è il cuore della catena di montaggio. Qui il Pilot costruisce letteralmente la ISO da zero, isolando il sistema e preparandolo per il boot su hardware sconosciuto.

### 1. Setup del Filesystem (`coa-fs-setup`)
*   **Azione:** `oa_mount_logic`
*   **Logica Pro:** Questa non è una semplice copia. Il Pilot espande dinamicamente questa direttiva in una sequenza di `oa_mkdir`, `oa_cp` e `oa_bind`. Crea un chroot pulito (`/home/eggs/liveroot`) e "proietta" i mount point vitali dell'host (`/dev`, `/proc`, `/sys`) all'interno dell'ambiente di staging per permettere l'esecuzione di comandi di sistema (come la generazione dell'initramfs) senza corrompere l'OS originale.

### 2. Bootloaders & UEFI Staging (`coa-bootloaders`)
*   **Azione:** `oa_shell` (Host)
*   **Logica Pro:** Prepara le fondamenta per l'avvio ibrido. 
    *   **Legacy BIOS:** Copia i binari standard `isolinux.bin` e i moduli necessari (`vesamenu.c32`, ecc.) per garantire la compatibilità con hardware datato.
    *   **UEFI (La magia FAT):** Costruisce un'immagine disco virtuale da 8MB (`efi.img`) formattata in FAT32 (`mkfs.vfat`). Utilizzando `mmd` e `mcopy` (dal pacchetto mtools), inietta al suo interno la struttura `/EFI/BOOT/` contenente il payload monolitico `BOOTX64.EFI` e un mini file `grub.cfg`. Questa immagine è ciò che permette a `xorriso` di rendere la ISO avviabile dai moderni firmware UEFI.

### 3. Gestione Identità e Sudo (`coa-identity` & `coa-enable-live`)
*   **Azione:** `oa_users` / `oa_shell` (Chroot)
*   **Logica Pro:** Implementa un approccio "Purge & Inject" in stile Yocto Project.
    *   `oa_users` ripulisce le tracce degli utenti originali per evitare leak di dati privati nella ISO.
    *   `coa-enable-live` inietta un file temporaneo in `/etc/sudoers.d/00-live` (con permessi restrittivi `0440`) garantendo all'utente live i poteri di amministratore senza password. Questo file verrà poi disintegrato in fase di installazione per chiudere la falla di sicurezza sul sistema finale.

### 4. Il Cuore del Boot: Initramfs (`coa-initrd`)
*   **Azione:** `oa_shell` (Chroot)
*   **Logica Pro:** Questa è la divergenza architetturale più marcata tra le distro. L'obiettivo è caricare i moduli necessari per avviare il sistema in memoria da un file SquashFS (overlayfs).
    *   **Debian:** `update-initramfs -u -k all` aggiorna le immagini standard.
    *   **Arch/Manjaro:** Sovrascrive temporaneamente `/etc/mkinitcpio.conf` per forzare gli hook necessari (`archiso_loop_mnt` per Arch, `miso_loop_mnt` per Manjaro) e compila con compressione `zstd`. Ripristina il file originale a operazione conclusa.
    *   **Fedora:** Utilizza `dracut` con i parametri vitali `--no-hostonly` (perché l'hardware su cui partirà la ISO non è quello su cui è stata buildata) e aggiunge il modulo `dmsquash-live`.

### 5. Estrazione del Kernel (`coa-kernel-copy` / `coa-copy-kernel`)
*   **Azione:** `oa_shell` (Host)
*   **Logica Pro:** Uno script dinamico esplora la cartella `/boot` del chroot per individuare il kernel corretto, ignorando versioni "fallback" o "rescue" (`grep -v fallback`, `tail -n 1`). I file estratti vengono rinominati in `vmlinuz` e `initrd.img` e piazzati nella root della ISO per mantenere una struttura standardizzata in tutti i menu di boot, a prescindere dalla distro di partenza.

### 6. Branding & Generazione Menu (`coa-live-menus`)
*   **Azione:** `oa_shell` (Host)
*   **Logica Pro:** Genera "al volo" i file di configurazione per tre sistemi di boot (GRUB EFI, GRUB Legacy, ISOLINUX).
    *   **Variabili d'Avvio:** Configura i parametri specifici per indicare al kernel dove trovare il filesystem compresso (`boot=live` per Debian, `archisobasedir=arch` per Arch, `root=live:LABEL=OA_LIVE` per Fedora).
    *   **Estetica:** Importa il font Unicode `font.pf2` per supportare tutti i caratteri e imposta un layout grafico. Un dettaglio fondamentale è il branding dinamico: estrae il nome della distro leggendo `PRETTY_NAME` da `/etc/os-release` ("Start Debian GNU/Linux 13...").
    *   **Trampolino EFI:** Genera un mini `grub.cfg` in `/EFI/BOOT/` che fa da ponte, dicendo al firmware di cercare la partizione con il file squashfs e di caricare il menu GRUB principale da lì.

### 7. Packaging: SquashFS & Ganci di Compatibilità (`coa-squashfs`)
*   **Azione:** `oa_shell` (Host)
*   **Logica Pro:** Comprime l'intero chroot in un singolo file `filesystem.squashfs` usando l'algoritmo `zstd` (livello 3, blocchi da 1M) ottimizzato per il multithreading (`-processors 4`).
*   **Ganci Architetturali:** Poiché gli hook di Arch, Manjaro e Fedora cercano percorsi hardcoded per montare il sistema live, il Pilot crea dei symlink strategici nella root della ISO (es. `LiveOS/squashfs.img` per Fedora o `arch/x86_64/airootfs.sfs` per Arch) per "ingannare" l'initramfs originale senza dover duplicare i pesanti file compressi.

### 8. Masterizzazione Finale (`coa-xorriso`)
*   **Azione:** `oa_shell` (Host)
*   **Logica Pro:** È il "Big Bang" della ISO. Il comando `xorriso` è settato per creare un'immagine "ibrida" (avviabile sia da USB che da DVD, sia in BIOS che in UEFI). Passa parametri critici come `-isohybrid-mbr` (usando il binario ISOLINUX) e `-eltorito-alt-boot` puntando all'immagine FAT `efi.img` creata nel passaggio 2.

---

## 🚀 PARTE 2: La Fase di Installazione (Finalizzazione Target)

La sezione `install:` entra in gioco alla fine del processo di Calamares. Quando tutti i file sono stati copiati sul disco dell'utente, il sistema è ancora "crudo". Il blocco `bootloader` trasforma quel mucchio di file in un OS funzionante.

### Lo Script Universale `oa-bootloader.sh`
*   **Azione:** `shell` (eseguito direttamente nel chroot del sistema target)
*   **Logica Pro:** Questo script rappresenta l'intelligenza adattiva dell'installatore.

1.  **Chiusura Falla di Sicurezza:** La prima istruzione è sempre `rm -f /etc/sudoers.d/00-live` per garantire che il sistema installato non mantenga l'utente amministratore senza password.
2.  **Rigenerazione Initramfs Target:** Esegue `update-initramfs`, `mkinitcpio` o `dracut` per rigenerare l'immagine di avvio basandosi sull'hardware reale della macchina ricevente.
3.  **Bivio Hardware (UEFI vs BIOS):**
    *   Tramite `[ -d /sys/firmware/efi ]` capta se l'utente ha avviato la ISO in modalità moderna o legacy.
    *   **Modalità UEFI:** Lancia `grub-install --target=x86_64-efi`, scrivendo nella partizione ESP (`/boot/efi`) senza toccare il MBR del disco.
    *   **Modalità BIOS:** Usa `grub-probe` per identificare la root reale (es. `/dev/sda`) e lancia un `grub-install` classico (`--target=i386-pc`).
4.  **Fix Specifico Fedora (BLS Disable):** Su macchine Fedora, disabilita il BootLoaderSpec (`GRUB_ENABLE_BLSCFG=false`). Questo obbliga il sistema a scrivere le `menuentry` esplicitamente all'interno di `grub.cfg`, garantendo un menu stabile e compatibile con l'estetica `oa`.
5.  **Generazione Configurazione:** Conclude lanciando `update-grub` o `grub-mkconfig` per applicare definitivamente il bootloader al disco.
