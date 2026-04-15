/*
 * src/actions/install_uefi.c
 * Remastering core: GRUB installation on physical hardware (Krill)
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

int install_uefi(OA_Context *ctx) {
    // 1. Lookup a cascata (disco e punto di mount)
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");

    if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(disk_node)) {
        LOG_ERR("pathLiveFs or run_command (disk) missing in install_uefi");
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Target disk or mount point missing.\n");
        return 1;
    }

    char target_root[PATH_SAFE];
    snprintf(target_root, sizeof(target_root), "%s/liveroot", pathLiveFs->valuestring);

    const char *disk = disk_node->valuestring;
    char cmd[CMD_MAX];
    int res = 0;

    printf("\033[1;34m[oa HATCH]\033[0m Installing GRUB bootloader to physical disk %s...\n", disk);

    // 2. Preparazione Ambiente (API Filesystems)
    // grub-install ha assoluto bisogno di /dev, /sys/firmware/efi e /proc per funzionare
    printf("  -> Mounting virtual filesystems for chroot...\n");
    snprintf(cmd, sizeof(cmd),
             "mount --bind /dev %s/dev && "
             "mount -t proc /proc %s/proc && "
             "mount -t sysfs /sys %s/sys && "
             "mount --bind /run %s/run",
             target_root, target_root, target_root, target_root);
             
    if (system(cmd) != 0) {
        LOG_ERR("Failed to mount virtual filesystems for UEFI installation");
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to setup chroot environment.\n");
        return 1;
    }

    // 3. Esecuzione NATIVA in chroot
    // Usiamo 'cova' come bootloader-id, questo sarà il nome visibile nel BIOS/UEFI della scheda madre
    printf("  -> Executing grub-install and mkconfig...\n");
    snprintf(cmd, sizeof(cmd),
             "chroot %s grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=cova --recheck > /dev/null && "
             "chroot %s grub-mkconfig -o /boot/grub/grub.cfg > /dev/null",
             target_root, target_root);
    
    res = system(cmd);

    // 4. Pulizia Tassativa
    // Usiamo ';' invece di '&&' perché vogliamo forzare l'umount di tutti anche se uno fallisce
    printf("  -> Unmounting virtual filesystems...\n");
    snprintf(cmd, sizeof(cmd),
             "umount %s/run ; umount %s/sys ; umount %s/proc ; umount %s/dev",
             target_root, target_root, target_root, target_root);
    system(cmd);

    // 5. Esito
    if (res == 0) {
        printf("\033[1;32m[SUCCESS]\033[0m GRUB installed successfully on %s.\n", disk);
        return 0;
    } else {
        LOG_ERR("GRUB installation failed on %s", disk);
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m GRUB installation failed. Check UEFI settings or partitions.\n");
        return 1;
    }
}