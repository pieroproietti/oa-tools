/*
 * src/actions/install_partition.c
 * Remastering core: Disk Partitioning
 * oa: eggs in my dialect🥚🥚
 */
#include "oa.h"

int install_partition(OA_Context *ctx) {
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Target disk missing in install_partition\n");
        return 1;
    }
    
    const char *disk = disk_node->valuestring;
    printf("\033[1;34m[oa]\033[0m Partitioning target disk: \033[1m%s\033[0m\n", disk);

    char cmd[512];

    printf("  -> Zapping old partition table...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --zap-all %s > /dev/null 2>&1", disk);
    if (system(cmd) != 0) return 1;

    // PARTIZIONE 1: BIOS Boot Partition (2MB)
    printf("  -> Creating BIOS Boot partition (2M)...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --new=1:0:+2M --typecode=1:ef02 --change-name=1:BIOSBOOT %s > /dev/null", disk);
    if (system(cmd) != 0) return 1;

    // PARTIZIONE 2: EFI System Partition (512MB)
    printf("  -> Creating EFI partition (512M)...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --new=2:0:+512M --typecode=2:ef00 --change-name=2:EFI %s > /dev/null", disk);
    if (system(cmd) != 0) return 1;

    // PARTIZIONE 3: ROOT (Spazio rimanente)
    printf("  -> Creating ROOT partition (remaining space)...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --new=3:0:0 --typecode=3:8300 --change-name=3:ROOT %s > /dev/null", disk);
    if (system(cmd) != 0) return 1;

    printf("  -> Syncing partition table with kernel...\n");
    snprintf(cmd, sizeof(cmd), "partprobe %s && udevadm settle", disk);
    system(cmd);

    printf("\033[1;32m[SUCCESS]\033[0m Disk %s successfully partitioned.\n", disk);
    return 0;
}