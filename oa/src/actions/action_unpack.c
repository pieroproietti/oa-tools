#include "../../include/oa.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Ricicliamo la funzione helper per i nomi delle partizioni
static void get_partition_path(const char *disk, int part_num, char *out_path, size_t max_len) {
    if (strstr(disk, "nvme") || strstr(disk, "mmcblk") || strstr(disk, "loop")) {
        snprintf(out_path, max_len, "%sp%d", disk, part_num);
    } else {
        snprintf(out_path, max_len, "%s%d", disk, part_num);
    }
}

int action_unpack(OA_Context *ctx) {
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        fprintf(stderr, "\033[1;31m[oa]\033[0m Target disk missing in action_unpack\n");
        return 1;
    }
    
    const char *disk = disk_node->valuestring;
    char efi_part[128];
    char root_part[128];
    
    get_partition_path(disk, 1, efi_part, sizeof(efi_part));
    get_partition_path(disk, 2, root_part, sizeof(root_part));

    // Andiamo a leggere dove dobbiamo montare il disco dal nodo "radice" del JSON
    cJSON *livefs_node = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    const char *mount_point = (cJSON_IsString(livefs_node) && livefs_node->valuestring) ? livefs_node->valuestring : "/mnt/krill-target";

    printf("\033[1;34m[oa]\033[0m Unpacking system to: \033[1m%s\033[0m\n", mount_point);
    char cmd[1024];

    // 1. Creazione Mount Point e Montaggio Root
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", mount_point);
    system(cmd);

    printf("  -> Mounting ROOT (%s) on %s...\n", root_part, mount_point);
    snprintf(cmd, sizeof(cmd), "mount %s %s", root_part, mount_point);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to mount root partition\n");
        return 1;
    }

    // 2. Creazione e Montaggio EFI
    printf("  -> Mounting EFI (%s) on %s/boot/efi...\n", efi_part, mount_point);
    snprintf(cmd, sizeof(cmd), "mkdir -p %s/boot/efi", mount_point);
    system(cmd);
    
    snprintf(cmd, sizeof(cmd), "mount %s %s/boot/efi", efi_part, mount_point);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to mount EFI partition\n");
        return 1;
    }

    // 3. Il Travaso (Rsync dal sistema corrente verso il nuovo disco)
    // Usiamo --info=progress2 per avere un output elegante sul terminale e non spammare righe
    printf("  -> Copying system data (this will take a while)...\n");
    const char *rsync_cmd = "rsync -aAX --info=progress2 "
                            "--exclude=/dev/* --exclude=/proc/* --exclude=/sys/* "
                            "--exclude=/tmp/* --exclude=/run/* --exclude=/mnt/* "
                            "--exclude=/media/* --exclude=/lost+found "
                            "--exclude=/home/eggs/* "
                            "/ %s/";
    
    snprintf(cmd, sizeof(cmd), rsync_cmd, mount_point);
    if (system(cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m System copy failed\n");
        return 1;
    }

    printf("\n\033[1;32m[SUCCESS]\033[0m System correctly unpacked.\n");
    return 0;
}
