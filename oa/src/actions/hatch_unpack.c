/*
 * src/actions/hatch_uefi.c
 * Remastering core: GRUB installation on physical hardware (Krill)
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

// Ricicliamo la funzione helper per i nomi delle partizioni
static void get_partition_path(const char *disk, int part_num, char *out_path, size_t max_len) {
    if (strstr(disk, "nvme") || strstr(disk, "mmcblk") || strstr(disk, "loop")) {
        snprintf(out_path, max_len, "%sp%d", disk, part_num);
    } else {
        snprintf(out_path, max_len, "%s%d", disk, part_num);
    }
}

int hatch_unpack(OA_Context *ctx) {
    // 1. Recupero parametri dal piano
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Target disk missing in hatch_unpack\n");
        return 1;
    }
    
    const char *disk = disk_node->valuestring;
    char efi_part[128], root_part[128];
    
    get_partition_path(disk, 1, efi_part, sizeof(efi_part));
    get_partition_path(disk, 2, root_part, sizeof(root_part));

    // 2. Definizione dei percorsi simmetrici
    cJSON *livefs_node = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    const char *nest_path = (cJSON_IsString(livefs_node) && livefs_node->valuestring) ? livefs_node->valuestring : "/mnt/krill";

    // Costruiamo la liveroot fisica per l'installazione
    char target_liveroot[PATH_SAFE];
    snprintf(target_liveroot, sizeof(target_liveroot), "%s/liveroot", nest_path);

    printf("\033[1;34m[oa HATCH]\033[0m Unpacking system to target: \033[1m%s\033[0m\n", target_liveroot);
    char cmd[CMD_MAX];

    // 3. Montaggio ROOT
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", target_liveroot);
    system(cmd);

    printf("  -> Mounting ROOT (%s) on %s...\n", root_part, target_liveroot);
    snprintf(cmd, sizeof(cmd), "mount %s %s", root_part, target_liveroot);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to mount root partition\n");
        return 1;
    }

    // 4. Montaggio EFI
    char efi_mount[PATH_SAFE];
    snprintf(efi_mount, sizeof(efi_mount), "%s/boot/efi", target_liveroot);

    printf("  -> Mounting EFI (%s) on %s...\n", efi_part, efi_mount);
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", efi_mount);
    system(cmd);
    
    snprintf(cmd, sizeof(cmd), "mount %s %s", efi_part, efi_mount);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to mount EFI partition\n");
        return 1;
    }

    // 5. Travaso dati via Rsync
    printf("  -> Copying system data (this will take a while)...\n");
    const char *rsync_cmd = "rsync -aAX --info=progress2 "
                            "--exclude=/dev/* --exclude=/proc/* --exclude=/sys/* "
                            "--exclude=/tmp/* --exclude=/run/* --exclude=/mnt/* "
                            "--exclude=/media/* --exclude=/lost+found "
                            "--exclude=%s " 
                            "/ %s/";
    
    snprintf(cmd, sizeof(cmd), rsync_cmd, nest_path, target_liveroot);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m System copy failed\n");
        return 1;
    }

    printf("\n\033[1;32m[SUCCESS]\033[0m System correctly unpacked into liveroot.\n");
    return 0;
}
