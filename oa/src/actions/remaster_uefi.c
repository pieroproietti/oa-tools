/*
 * src/actions/remaster_uefi.c
 * Remastering core: GRUB installation on physical hardware (Krill)
 * oa: uova nel mio dialetto 🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

int remaster_uefi(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    cJSON *bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->task, "bootloaders_path");
    if (!bootloaders_path) bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->root, "bootloaders_path");

    char efi_dir[PATH_SAFE], grub_dir[PATH_SAFE], grub_cfg_dir[PATH_SAFE];
    snprintf(efi_dir, PATH_SAFE, "%s/iso/EFI/BOOT", pathLiveFs->valuestring);
    snprintf(grub_dir, PATH_SAFE, "%s/iso/boot/grub/x86_64-efi", pathLiveFs->valuestring);
    snprintf(grub_cfg_dir, PATH_SAFE, "%s/iso/boot/grub", pathLiveFs->valuestring);

    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s %s", efi_dir, grub_dir);
    system(cmd);

    printf("\033[1;34m[oa UEFI]\033[0m Preparing UEFI boot directories...\n");

    // 1. Logica di selezione della sorgente con Prefisso Unificato 
    const char *prefix = (cJSON_IsString(bootloaders_path) && strlen(bootloaders_path->valuestring) > 0) 
                         ? bootloaders_path->valuestring 
                         : "";

    char efi_src[PATH_SAFE] = "";
    char grub_mods_src[PATH_SAFE] = "";

    if (prefix[0] != '\0') {
        // Percorsi basati sulla struttura del tarball (/tmp/coa/bootloaders/)
        // Puntiamo alla sottocartella monolithic dove risiede grubx64.efi
        snprintf(efi_src, PATH_SAFE, "%s/grub/x86_64-efi/monolithic/grubx64.efi", prefix);
        snprintf(grub_mods_src, PATH_SAFE, "%s/grub/x86_64-efi", prefix);
        printf("\033[1;34m[oa UEFI]\033[0m Using external bootloaders from: %s\n", prefix);
    } else {
        // Fallback standard host Debian
        if (access("/usr/lib/grub/x86_64-efi/monolithic/grubx64.efi", F_OK) == 0) {
            strncpy(efi_src, "/usr/lib/grub/x86_64-efi/monolithic/grubx64.efi", PATH_SAFE);
        } else if (access("/boot/efi/EFI/debian/grubx64.efi", F_OK) == 0) {
            strncpy(efi_src, "/boot/efi/EFI/debian/grubx64.efi", PATH_SAFE);
        }
        strncpy(grub_mods_src, "/usr/lib/grub/x86_64-efi", PATH_SAFE);
    }

    // 2. Copia del payload EFI (bootx64.efi)
    if (access(efi_src, F_OK) == 0) {
        snprintf(cmd, sizeof(cmd), "cp %s %s/bootx64.efi", efi_src, efi_dir);
        system(cmd);
        LOG_INFO("Extracted UEFI bootloader from %s", efi_src);
    } else {
        LOG_WARN("No EFI payload found. UEFI boot might fail.");
        printf("\033[1;33m[oa UEFI]\033[0m Warning: No EFI payload found.\n");
    }

    // 3. Copia dei moduli GRUB (*.mod, *.lst, ecc.)
    if (access(grub_mods_src, F_OK) == 0) {
        snprintf(cmd, sizeof(cmd), "cp -r %s/* %s/ 2>/dev/null || true", grub_mods_src, grub_dir);
        system(cmd);
        LOG_INFO("Extracted GRUB modules from %s", grub_mods_src);
    } else {
        LOG_WARN("GRUB modules not found at %s", grub_mods_src);
    }

    // 4. Generazione automatica di grub.cfg (MENU PRINCIPALE)
    char cfg_path[PATH_SAFE];
    snprintf(cfg_path, PATH_SAFE, "%s/grub.cfg", grub_cfg_dir);

    FILE *f = fopen(cfg_path, "w");
    if (f) {
        fprintf(f, 
            "search --no-floppy --set=root --file /live/vmlinuz\n\n"
            "set default=\"0\"\n"
            "set timeout=10\n\n"
            "menuentry \"oa Live System (UEFI)\" {\n"
            "    echo \"Loading kernel...\"\n"
            "    linux /live/vmlinuz boot=live components quiet splash\n"
            "    echo \"Loading initial ramdisk...\"\n"
            "    initrd /live/initrd.img\n"
            "}\n"
        );
        fclose(f);
        printf("\033[1;32m[oa UEFI]\033[0m Main grub.cfg generated.\n");
    }

    // 5. Generazione del grub.cfg TRAMPOLINO
    char efi_cfg_path[PATH_SAFE];
    snprintf(efi_cfg_path, PATH_SAFE, "%s/grub.cfg", efi_dir);

    FILE *f_efi = fopen(efi_cfg_path, "w");
    if (f_efi) {
        fprintf(f_efi, 
            "search --no-floppy --set=root --file /boot/grub/grub.cfg\n"
            "set prefix=($root)/boot/grub\n"
            "configfile /boot/grub/grub.cfg\n"
        );
        fclose(f_efi);
        printf("\033[1;32m[oa UEFI]\033[0m EFI trampoline grub.cfg generated.\n");
    }

    return 0;
}