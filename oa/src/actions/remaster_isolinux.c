/*
 * src/actions/remaster_isolinux.c
 * Corretto secondo la struttura reale del tarball bootloaders
 */
#include "oa.h"

int remaster_isolinux(OA_Context *ctx) {
    // Lookup dei percorsi
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    cJSON *bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->task, "bootloaders_path");
    if (!bootloaders_path) bootloaders_path = cJSON_GetObjectItemCaseSensitive(ctx->root, "bootloaders_path");

    cJSON *params_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "boot_params");
    if (!params_obj) params_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "boot_params");
    
    cJSON *family_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "family");

    const char *boot_params = cJSON_IsString(params_obj) ? params_obj->valuestring : "boot=live components quiet splash";

    char isolinux_dir[PATH_SAFE];
    snprintf(isolinux_dir, PATH_SAFE, "%s/iso/isolinux", pathLiveFs->valuestring);

    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", isolinux_dir);
    system(cmd);

    const char *prefix = (cJSON_IsString(bootloaders_path) && strlen(bootloaders_path->valuestring) > 0) 
                         ? bootloaders_path->valuestring : "";

    // --- CORREZIONE COPIA SECONDO TREE ---
    if (prefix[0] != '\0') {
        printf("\033[1;34m[oa ISOLINUX]\033[0m Sourcing from tree: %s\n", prefix);

        // 1. isolinux.bin sta in /ISOLINUX/ 
        snprintf(cmd, sizeof(cmd), "cp %s/ISOLINUX/isolinux.bin %s/", prefix, isolinux_dir);
        system(cmd);

        // 2. I file .c32 core (ldlinux, vesamenu, etc) stanno in /syslinux/modules/bios/ [cite: 24, 27]
        // Usiamo sh -c per garantire l'espansione del jolly '*'
        snprintf(cmd, sizeof(cmd), "sh -c 'cp %s/syslinux/modules/bios/*.c32 %s/'", prefix, isolinux_dir);
        if (system(cmd) != 0) {
            fprintf(stderr, "[ERROR] Fallita la copia dei moduli .c32 da syslinux/modules/bios/\n");
            return 1;
        }
    } else {
        // Fallback locale host
        snprintf(cmd, sizeof(cmd), "cp /usr/lib/syslinux/modules/bios/*.c32 %s/ 2>/dev/null || cp /usr/lib/syslinux/bios/*.c32 %s/", isolinux_dir, isolinux_dir);
        system(cmd);
    }

    // --- TRUCCO ARCH ---
    if (family_obj && strcmp(family_obj->valuestring, "archlinux") == 0) {
        char arch_path[PATH_SAFE];
        snprintf(arch_path, sizeof(arch_path), "%s/iso/arch/x86_64", pathLiveFs->valuestring);
        snprintf(cmd, sizeof(cmd), "mkdir -p %s && ln -sf ../../live/filesystem.squashfs %s/airootfs.sfs", 
                 arch_path, arch_path);
        system(cmd);
    }

    // --- CONFIGURAZIONE ---
    char cfg_path[PATH_SAFE];
    snprintf(cfg_path, PATH_SAFE, "%s/isolinux.cfg", isolinux_dir);
    FILE *f = fopen(cfg_path, "w");
    if (f) {
        fprintf(f, "UI vesamenu.c32\n"
                   "PROMPT 0\n"
                   "TIMEOUT 50\n"
                   "DEFAULT live\n"
                   "MENU TITLE oa Tools\n"
                   "LABEL live\n"
                   "  MENU LABEL Boot Live System\n"
                   "  KERNEL /live/vmlinuz\n"
                   "  APPEND initrd=/live/initrd.img %s\n", boot_params);
        fclose(f);
    }

    return 0;
}
