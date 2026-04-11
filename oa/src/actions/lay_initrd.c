/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/

#include "oa.h"

/**
 * @brief Helper locale per sostituire i placeholder nel comando initrd
 */
static void build_initrd_command(char *dest, const char *tpl, const char *out, const char *ver) {
    char tmp[4096];
    const char *p;

    // Sostituisce {{out}}
    if ((p = strstr(tpl, "{{out}}"))) {
        size_t len = p - tpl;
        strncpy(tmp, tpl, len);
        tmp[len] = '\0';
        strcat(tmp, out);
        strcat(tmp, p + 7);
        tpl = tmp;
    }
    // Sostituisce {{ver}}
    if ((p = strstr(tpl, "{{ver}}"))) {
        size_t len = p - tpl;
        char final[4096];
        strncpy(final, tpl, len);
        final[len] = '\0';
        strcat(final, ver);
        strcat(final, p + 7);
        strcpy(dest, final);
    } else {
        strcpy(dest, tpl);
    }
}

/**
 * @brief Genera l'initrd (Initial RAM Disk) usando il contesto OA
 */
int lay_initrd(OA_Context *ctx) {
    // 1. Lookup a cascata (Locale > Globale)
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(ctx->task, "initrd_cmd");
    if (!initrd_cmd_tpl) initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(ctx->root, "initrd_cmd");

    if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(initrd_cmd_tpl)) {
        fprintf(stderr, "Error: initrd_cmd or pathLiveFs missing\n");
        return 1;
    }

    char live_dir[PATH_SAFE], final_cmd[4096], initrd_out[PATH_SAFE];
    snprintf(live_dir, PATH_SAFE, "%s/iso/live", pathLiveFs->valuestring);
    snprintf(initrd_out, PATH_SAFE, "%s/initrd.img", live_dir);

    // 2. Rilevamento kernel host
    struct utsname buffer;
    uname(&buffer);
    char *kversion = buffer.release;

    // 3. Costruzione comando
    build_initrd_command(final_cmd, initrd_cmd_tpl->valuestring, initrd_out, kversion);

    printf("\033[1;34m[oa]\033[0m Generating initrd: %s\n", final_cmd);
    
    // Assicuriamoci che la cartella esista
    char mkdir_cmd[PATH_SAFE];
    snprintf(mkdir_cmd, PATH_SAFE, "mkdir -p %s", live_dir);
    system(mkdir_cmd);

    if (system(final_cmd) != 0) {
        fprintf(stderr, "Error: initrd generation failed\n");
        return 1;
    }

    return 0;
}
