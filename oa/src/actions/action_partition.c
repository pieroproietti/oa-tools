#include "../../include/oa.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int action_partition(OA_Context *ctx) {
    // Leggiamo il JSON direttamente dal contesto (ctx->task)
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Target disk missing in action_partition\n");
        return 1; // 1 = Errore
    }
    
    const char *disk = disk_node->valuestring;
    printf("\033[1;34m[oa]\033[0m Partitioning target disk: \033[1m%s\033[0m\n", disk);

    char cmd[512];

    printf("  -> Zapping old partition table...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --zap-all %s > /dev/null 2>&1", disk);
    if (system(cmd) != 0) return 1;

    printf("  -> Creating EFI partition (512M)...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --new=1:0:+512M --typecode=1:ef00 --change-name=1:EFI %s > /dev/null", disk);
    if (system(cmd) != 0) return 1;

    printf("  -> Creating ROOT partition (remaining space)...\n");
    snprintf(cmd, sizeof(cmd), "sgdisk --new=2:0:0 --typecode=2:8300 --change-name=2:ROOT %s > /dev/null", disk);
    if (system(cmd) != 0) return 1;

    printf("  -> Syncing partition table with kernel...\n");
    snprintf(cmd, sizeof(cmd), "partprobe %s && udevadm settle", disk);
    system(cmd); // udevadm è critico ma non fermiamo l'esecuzione se partprobe si lamenta

    printf("\033[1;32m[SUCCESS]\033[0m Disk %s successfully partitioned.\n", disk);
    return 0; // 0 = Successo
}
