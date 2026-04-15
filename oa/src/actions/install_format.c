/*
 * src/actions/install_uefi.c
 * Remastering core: GRUB installation on physical hardware (Krill)
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

static void get_partition_path(const char *disk, int part_num, char *out_path, size_t max_len) {
    if (strstr(disk, "nvme") || strstr(disk, "mmcblk") || strstr(disk, "loop")) {
        snprintf(out_path, max_len, "%sp%d", disk, part_num);
    } else {
        snprintf(out_path, max_len, "%s%d", disk, part_num);
    }
}

int install_format(OA_Context *ctx) {
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Target disk missing in install_format\n");
        return 1; 
    }
    
    const char *disk = disk_node->valuestring;
    char efi_part[128];
    char root_part[128];
    
    // AGGIORNATO: EFI è la 2, ROOT è la 3
    get_partition_path(disk, 2, efi_part, sizeof(efi_part));
    get_partition_path(disk, 3, root_part, sizeof(root_part));

    printf("\033[1;34m[oa]\033[0m Formatting partitions on: \033[1m%s\033[0m\n", disk);
    
    char cmd[512];
    printf("  -> Formatting EFI (%s) as FAT32...\n", efi_part);
    snprintf(cmd, sizeof(cmd), "mkfs.vfat -F32 %s > /dev/null", efi_part);
    if (system(cmd) != 0) return 1;

    printf("  -> Formatting ROOT (%s) as EXT4...\n", root_part);
    snprintf(cmd, sizeof(cmd), "mkfs.ext4 -F %s > /dev/null 2>&1", root_part);
    if (system(cmd) != 0) return 1;

    printf("\033[1;32m[SUCCESS]\033[0m Partitions successfully formatted.\n");
    return 0;
}