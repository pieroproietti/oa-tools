#include "../../include/oa.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/**
 * @brief Helper per calcolare il nome esatto della partizione
 * Gestisce la differenza tra dischi SATA (/dev/sdb1) e NVMe/MMC (/dev/nvme0n1p1)
 */
static void get_partition_path(const char *disk, int part_num, char *out_path, size_t max_len) {
    if (strstr(disk, "nvme") || strstr(disk, "mmcblk") || strstr(disk, "loop")) {
        snprintf(out_path, max_len, "%sp%d", disk, part_num);
    } else {
        snprintf(out_path, max_len, "%s%d", disk, part_num);
    }
}

/**
 * @brief Formatta la partizione 1 in FAT32 (EFI) e la 2 in EXT4 (Root)
 */
int action_format_ext4(OA_Context *ctx) {
    // Leggiamo il disco bersaglio direttamente dal task JSON nel contesto
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Target disk missing in action_format_ext4\n");
        return 1; // 1 = Errore
    }
    
    const char *disk = disk_node->valuestring;
    char efi_part[128];
    char root_part[128];
    
    // Calcoliamo i device file delle partizioni
    get_partition_path(disk, 1, efi_part, sizeof(efi_part));
    get_partition_path(disk, 2, root_part, sizeof(root_part));

    printf("\033[1;34m[oa]\033[0m Formatting partitions on: \033[1m%s\033[0m\n", disk);
    
    char cmd[512];

    // 1. Format EFI come FAT32 (richiede mkfs.vfat / dosfstools)
    printf("  -> Formatting EFI (%s) as FAT32...\n", efi_part);
    snprintf(cmd, sizeof(cmd), "mkfs.vfat -F32 %s > /dev/null", efi_part);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to format EFI partition\n");
        return 1;
    }

    // 2. Format ROOT come EXT4 (forzato con -F per evitare prompt interattivi)
    printf("  -> Formatting ROOT (%s) as EXT4...\n", root_part);
    snprintf(cmd, sizeof(cmd), "mkfs.ext4 -F %s > /dev/null 2>&1", root_part);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to format ROOT partition\n");
        return 1;
    }

    printf("\033[1;32m[SUCCESS]\033[0m Partitions successfully formatted.\n");
    return 0; // 0 = Successo
}
