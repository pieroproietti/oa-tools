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
int remaster_squash(OA_Context *ctx) {
    // 1. Lookup dei parametri principali
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *comp = cJSON_GetObjectItemCaseSensitive(ctx->task, "compression");
    if (!comp) comp = cJSON_GetObjectItemCaseSensitive(ctx->root, "compression");

    cJSON *comp_lvl = cJSON_GetObjectItemCaseSensitive(ctx->task, "compression_level");
    if (!comp_lvl) comp_lvl = cJSON_GetObjectItemCaseSensitive(ctx->root, "compression_level");

    // IL NUOVO FILE DELLE ESCLUSIONI GENERATO DA COA
    cJSON *exclude_file = cJSON_GetObjectItemCaseSensitive(ctx->task, "exclude_list");
    if (!exclude_file) exclude_file = cJSON_GetObjectItemCaseSensitive(ctx->root, "exclude_list");

    if (!cJSON_IsString(pathLiveFs)) return 1;

    // 2. Setup Base
    long nprocs = sysconf(_SC_NPROCESSORS_ONLN);
    int level = cJSON_IsNumber(comp_lvl) ? comp_lvl->valueint : 3;
    const char *comp_str = cJSON_IsString(comp) ? comp->valuestring : "zstd";

    char liveroot[PATH_SAFE], squash_out[PATH_SAFE];
    snprintf(liveroot, PATH_SAFE, "%s/liveroot", pathLiveFs->valuestring);
    snprintf(squash_out, PATH_SAFE, "%s/iso/live/filesystem.squashfs", pathLiveFs->valuestring);

    // 3. Esclusioni di Sopravvivenza (Hardcoded minime per la sicurezza del Kernel)
    char session_excludes[4096] = "";
    const char *survival_excludes[] = {
        "proc/*", "sys/*", "dev/*", "run/*", "tmp/*"
    };
    for (size_t i = 0; i < 5; i++) {
        append_eggs_exclusion(session_excludes, 4096, survival_excludes[i]);
    }

    // 4. Composizione Comando
    char cmd[CMD_MAX], comp_opts[256] = "";
    if (strcmp(comp_str, "zstd") == 0) snprintf(comp_opts, 256, "-Xcompression-level %d", level);

    snprintf(cmd, sizeof(cmd), "mksquashfs %s %s -comp %s %s -processors %ld -b 1M -noappend -wildcards", 
             liveroot, squash_out, comp_str, comp_opts, nprocs);

    // 5. Iniezione del file custom (generato dalla Mente in Go)
    if (cJSON_IsString(exclude_file) && access(exclude_file->valuestring, F_OK) == 0) {
        snprintf(cmd + strlen(cmd), CMD_MAX - strlen(cmd), " -ef %s", exclude_file->valuestring);
        printf("\033[1;34m[oa SQUASH]\033[0m Applying dynamic exclude list: %s\n", exclude_file->valuestring);
    }

    // 6. Iniezione stringa sopravvivenza
    if (strlen(session_excludes) > 0) {
        snprintf(cmd + strlen(cmd), CMD_MAX - strlen(cmd), " -e%s", session_excludes);
    }

    printf("\n\033[1;34m[oa SQUASH]\033[0m Cores: %ld | Lvl: %d | Comp: %s\n", nprocs, level, comp_str);
    return system(cmd);
}
