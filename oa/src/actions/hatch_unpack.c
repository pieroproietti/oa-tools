/*
 * src/actions/hatch_unpack.c
 * Remastering core: Unpack squashfs to physical disk
 * oa: eggs in my dialect🥚🥚
 */
#include "oa.h"

static void get_partition_path(const char *disk, int part_num, char *out_path, size_t max_len) {
    if (strstr(disk, "nvme") || strstr(disk, "mmcblk") || strstr(disk, "loop")) {
        snprintf(out_path, max_len, "%sp%d", disk, part_num);
    } else {
        snprintf(out_path, max_len, "%s%d", disk, part_num);
    }
}

int hatch_unpack(OA_Context *ctx) {
    cJSON *disk_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    cJSON *args_node = cJSON_GetObjectItemCaseSensitive(ctx->task, "args"); // <--- LEGGIAMO GLI ARGS DAL JSON

    if (!cJSON_IsString(disk_node) || (disk_node->valuestring == NULL)) {
        LOG_ERR("Target disk missing in hatch_unpack");
        return 1;
    }
    
    // Sicurezza: controlliamo che Go ci abbia effettivamente passato il percorso
    if (!cJSON_IsArray(args_node) || cJSON_GetArraySize(args_node) == 0) {
        LOG_ERR("Squashfs pristine path not provided in args array by coa");
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Pristine squashfs path missing in flight plan.\n");
        return 1;
    }

    const char *disk = disk_node->valuestring;
    const char *squash_file = cJSON_GetArrayItem(args_node, 0)->valuestring; // Estraiamo il percorso
    char efi_part[128], root_part[128];
    
    // EFI è la 2, ROOT è la 3
    get_partition_path(disk, 2, efi_part, sizeof(efi_part));
    get_partition_path(disk, 3, root_part, sizeof(root_part));

    cJSON *livefs_node = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    const char *nest_path = (cJSON_IsString(livefs_node) && livefs_node->valuestring) ? livefs_node->valuestring : "/mnt/krill";
    
    char target_liveroot[PATH_SAFE];
    snprintf(target_liveroot, sizeof(target_liveroot), "%s/liveroot", nest_path);
    
    printf("\033[1;34m[oa HATCH]\033[0m Sourcing pristine system image from: \033[1m%s\033[0m\n", squash_file);
    
    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", target_liveroot);
    system(cmd);
    
    // 1. Montiamo SOLO la ROOT
    printf("  -> Mounting ROOT (%s) on %s...\n", root_part, target_liveroot);
    snprintf(cmd, sizeof(cmd), "mount %s %s", root_part, target_liveroot);
    if (system(cmd) != 0) return 1;

    // 2. UNPACK dello Squashfs con mksquashfs-tools
    printf("  -> Unpacking pure system...\n");
    snprintf(cmd, sizeof(cmd), "unsquashfs -f -d %s %s", target_liveroot, squash_file);
    if (system(cmd) != 0) {
        LOG_ERR("Unsquashfs extraction failed on %s", squash_file);
        return 1;
    }

    // 3. Creiamo il mountpoint EFI nel sistema appena scompattato e lo montiamo
    char efi_mount[PATH_SAFE];
    snprintf(efi_mount, sizeof(efi_mount), "%s/boot/efi", target_liveroot);
    printf("  -> Mounting EFI (%s) on %s...\n", efi_part, efi_mount);
    
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", efi_mount);
    system(cmd);
    
    snprintf(cmd, sizeof(cmd), "mount %s %s", efi_part, efi_mount);
    if (system(cmd) != 0) return 1;

    printf("\033[1;32m[SUCCESS]\033[0m Pristine system correctly unpacked and ready for setup.\n");
    return 0;
}
