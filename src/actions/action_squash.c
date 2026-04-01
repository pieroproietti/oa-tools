/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"
#include "helpers.h"

/**
 * @brief Crea il filesystem compresso SquashFS
 */
int action_squash(OA_Context *ctx) {
    // Lookup intelligente dei parametri
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *comp = cJSON_GetObjectItemCaseSensitive(ctx->task, "compression");
    if (!comp) comp = cJSON_GetObjectItemCaseSensitive(ctx->root, "compression");

    cJSON *comp_lvl = cJSON_GetObjectItemCaseSensitive(ctx->task, "compression_level");
    if (!comp_lvl) comp_lvl = cJSON_GetObjectItemCaseSensitive(ctx->root, "compression_level");

    cJSON *exclude_file = cJSON_GetObjectItemCaseSensitive(ctx->task, "exclude_list");
    if (!exclude_file) exclude_file = cJSON_GetObjectItemCaseSensitive(ctx->root, "exclude_list");

    cJSON *mode_item = cJSON_GetObjectItemCaseSensitive(ctx->task, "mode");
    if (!mode_item) mode_item = cJSON_GetObjectItemCaseSensitive(ctx->root, "mode");

    if (!cJSON_IsString(pathLiveFs)) return 1;

    // Da qui in poi la logica rimane la stessa, usando i puntatori estratti sopra
    const char *mode = (cJSON_IsString(mode_item)) ? mode_item->valuestring : "";
    char final_exclude_path[PATH_SAFE] = "";
    if (cJSON_IsString(exclude_file) && access(exclude_file->valuestring, F_OK) == 0) {
        strncpy(final_exclude_path, exclude_file->valuestring, PATH_MAX);
    } else if (access("/usr/share/oa/exclusion.list", F_OK) == 0) {
        strncpy(final_exclude_path, "/usr/share/oa/exclusion.list", PATH_MAX);
    }

    long nprocs = sysconf(_SC_NPROCESSORS_ONLN);
    int level = cJSON_IsNumber(comp_lvl) ? comp_lvl->valueint : 3;
    const char *comp_str = cJSON_IsString(comp) ? comp->valuestring : "zstd";

    char liveroot[PATH_SAFE], squash_out[PATH_SAFE];
    snprintf(liveroot, PATH_SAFE, "%s/liveroot", pathLiveFs->valuestring);
    snprintf(squash_out, PATH_SAFE, "%s/iso/live/filesystem.squashfs", pathLiveFs->valuestring);

    char session_excludes[4096] = "";
    const char *fexcludes[] = {
        "boot/efi/EFI", "boot/loader/entries/", "etc/fstab", "var/lib/docker/",
        "proc/*", "sys/*", "dev/*", "run/*", "tmp/*"
    };
    for (size_t i = 0; i < 9; i++) append_eggs_exclusion(session_excludes, 4096, fexcludes[i]);

    if (strcmp(mode, "clone") != 0) {
        // append_eggs_exclusion(session_excludes, 4096, "home/*");
        append_eggs_exclusion(session_excludes, 4096, "root/*");
    }

    char cmd[8192], comp_opts[256] = "";
    if (strcmp(comp_str, "zstd") == 0) snprintf(comp_opts, 256, "-Xcompression-level %d", level);

    snprintf(cmd, sizeof(cmd), "mksquashfs %s %s -comp %s %s -processors %ld -b 1M -noappend -wildcards", 
             liveroot, squash_out, comp_str, comp_opts, nprocs);

    if (strlen(final_exclude_path) > 0) {
        snprintf(cmd + strlen(cmd), CMD_MAX - strlen(cmd), " -ef %s", final_exclude_path);
    }
    if (strlen(session_excludes) > 0) {
        snprintf(cmd + strlen(cmd), 8192 - strlen(cmd), " -e%s", session_excludes);
    }

    printf("\n\033[1;34m[oa Turbo Squash]\033[0m Cores: %ld | Lvl: %d | Mode: %s\n", nprocs, level, mode);
    return system(cmd);
}

